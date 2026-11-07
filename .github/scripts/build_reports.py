#!/usr/bin/env python3
"""Собирает HTML-отчёты (coverage, tests, lint) для GitHub Pages.

Вход (лежит в reports/, скачано из artifacts CI):
  reports/lint/lint.txt, reports/lint/lint.status   вывод и код возврата golangci-lint
  reports/test/test.json, reports/test/test.status  вывод `go test -json` и код возврата
  reports/test/coverage.out                         профиль покрытия (уже отфильтрованный)
  reports/test/coverage.html                        вывод `go tool cover -html`

Выход (каталог --site):
  index.html            сводка
  coverage/index.html   покрытие по пакетам, coverage/full.html — подсветка кода
  tests/index.html      результаты тестов по пакетам
  lint/index.html       вывод линтера
  summary.json          числа для бейджей

Если передан --readme, обновляет бейдж Coverage в README.
"""
from __future__ import annotations

import argparse
import collections
import datetime as dt
import html
import json
import os
import pathlib
import re
import shutil

CSS = """
:root{--bg:#1b1e23;--card:#30363e;--line:#444b55;--fg:#eceff2;--dim:#8b929c;
--accent:#38c0e6;--ok:#3fb950;--bad:#f85149;--warn:#d29922}
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--fg);
font:15px/1.5 Inter,-apple-system,"Helvetica Neue",Arial,sans-serif}
main{max-width:1000px;margin:0 auto;padding:32px 20px 64px}
nav{display:flex;gap:6px;flex-wrap:wrap;margin-bottom:28px}
nav a{color:var(--dim);text-decoration:none;padding:6px 14px;border-radius:999px;border:1px solid var(--line)}
nav a:hover,nav a.on{color:var(--fg);border-color:var(--accent)}
h1{font-size:28px;margin:0 0 4px;letter-spacing:-.5px}
.meta{color:var(--dim);font-size:13px;margin-bottom:24px}
.meta code{font-family:"SF Mono",Menlo,monospace}
.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(260px,1fr));gap:16px}
.card{display:block;background:var(--card);border:1px solid var(--line);border-radius:14px;
padding:20px;color:inherit;text-decoration:none}
a.card:hover{border-color:var(--accent)}
.card h2{margin:0 0 8px;font-size:13px;font-weight:600;color:var(--dim);letter-spacing:1.5px;text-transform:uppercase}
.big{font-size:40px;font-weight:700;letter-spacing:-1px}
.ok{color:var(--ok)}.bad{color:var(--bad)}.warn{color:var(--warn)}
.sub{color:var(--dim);font-size:13px}
table{width:100%;border-collapse:collapse;margin:16px 0;background:var(--card);
border:1px solid var(--line);border-radius:14px;overflow:hidden}
th,td{text-align:left;padding:10px 14px;border-bottom:1px solid var(--line)}
th{font-size:12px;color:var(--dim);font-weight:600;letter-spacing:1px;text-transform:uppercase}
tr:last-child td{border-bottom:0}
td.n,th.n{text-align:right;font-variant-numeric:tabular-nums}
td.p{font-family:"SF Mono",Menlo,monospace;font-size:13px}
.bar{height:6px;border-radius:3px;background:var(--line);min-width:80px}
.bar i{display:block;height:100%;border-radius:3px;background:var(--ok)}
pre{background:#14171b;border:1px solid var(--line);border-radius:12px;padding:16px;overflow:auto;
font:12.5px/1.5 "SF Mono",Menlo,monospace;color:#cfd4da;white-space:pre-wrap}
.banner{padding:12px 16px;border-radius:12px;border:1px solid var(--line);margin-bottom:20px;font-weight:600}
.banner.ok{border-color:var(--ok)}.banner.bad{border-color:var(--bad)}
details{margin:10px 0}summary{cursor:pointer;font-family:"SF Mono",Menlo,monospace;font-size:13px}
"""


def esc(s: str) -> str:
    return html.escape(s, quote=False)


def read(path: pathlib.Path) -> str | None:
    return path.read_text(encoding="utf-8", errors="replace") if path.exists() else None


def module_name() -> str:
    gomod = read(pathlib.Path("go.mod")) or ""
    m = re.search(r"^module\s+(\S+)", gomod, re.M)
    return m.group(1) if m else ""


def meta_line() -> str:
    sha = os.environ.get("GITHUB_SHA", "local")[:7]
    ref = os.environ.get("GITHUB_REF_NAME", "local")
    now = dt.datetime.now(dt.timezone.utc).strftime("%Y-%m-%d %H:%M UTC")
    return f'{now} · ветка <code>{esc(ref)}</code> · коммит <code>{esc(sha)}</code>'


def page(title: str, body: str, active: str, root: str) -> str:
    items = [("index", "Сводка", f"{root}index.html"), ("coverage", "Coverage", f"{root}coverage/"),
             ("tests", "Tests", f"{root}tests/"), ("lint", "Lint", f"{root}lint/")]
    nav = "".join(f'<a href="{href}"{" class=on" if key == active else ""}>{label}</a>'
                  for key, label, href in items)
    return (f'<!doctype html><html lang="ru"><head><meta charset="utf-8">'
            f'<meta name="viewport" content="width=device-width,initial-scale=1">'
            f'<title>{esc(title)}</title><style>{CSS}</style></head><body><main>'
            f'<nav>{nav}</nav><h1>{esc(title)}</h1><div class="meta">{meta_line()}</div>{body}'
            f'</main></body></html>')


def write(path: pathlib.Path, content: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content, encoding="utf-8")


# ---------------------------------------------------------------- coverage

def parse_coverage(text: str, mod: str):
    # При -coverpkg один блок встречается в профиле по разу на каждый тестовый
    # бинарник; как и `go tool cover`, склеиваем дубликаты, суммируя счётчики.
    blocks: dict[tuple[str, str], list[int]] = {}
    for line in text.splitlines():
        m = re.match(r"^(.+?):(\d+\.\d+,\d+\.\d+) (\d+) (\d+)$", line)
        if not m:
            continue
        b = blocks.setdefault((m.group(1), m.group(2)), [int(m.group(3)), 0])
        b[1] += int(m.group(4))
    total = covered = 0
    pkgs: dict[str, list[int]] = collections.defaultdict(lambda: [0, 0])
    for (file, _), (stmts, count) in blocks.items():
        pkg = file.rsplit("/", 1)[0]
        if mod and pkg.startswith(mod + "/"):
            pkg = pkg[len(mod) + 1:]
        pkgs[pkg][0] += stmts
        total += stmts
        if count > 0:
            pkgs[pkg][1] += stmts
            covered += stmts
    pct = round(100.0 * covered / total, 1) if total else 0.0
    return pct, pkgs


def pct_class(p: float) -> str:
    return "ok" if p >= 70 else "warn" if p >= 40 else "bad"


def build_coverage(src: pathlib.Path, site: pathlib.Path, mod: str):
    text = read(src / "test" / "coverage.out")
    if text is None:
        write(site / "coverage" / "index.html",
              page("Coverage", '<div class="banner bad">Профиль покрытия не получен — смотрите job test.</div>',
                   "coverage", "../"))
        return None
    pct, pkgs = parse_coverage(text, mod)
    rows = []
    for pkg, (n, c) in sorted(pkgs.items()):
        p = round(100.0 * c / n, 1) if n else 0.0
        rows.append(f'<tr><td class="p">{esc(pkg)}</td><td class="n">{c}/{n}</td>'
                    f'<td class="n {pct_class(p)}">{p}%</td>'
                    f'<td><div class="bar"><i style="width:{p}%"></i></div></td></tr>')
    full = src / "test" / "coverage.html"
    link = ""
    if full.exists():
        (site / "coverage").mkdir(parents=True, exist_ok=True)
        shutil.copy(full, site / "coverage" / "full.html")
        link = '<p><a style="color:var(--accent)" href="full.html">Открыть подсветку кода →</a></p>'
    body = (f'<div class="card" style="margin-bottom:16px"><h2>Покрытие бизнес-логики</h2>'
            f'<div class="big {pct_class(pct)}">{pct}%</div>'
            f'<div class="sub">пакеты internal/, без fakes и test-хелперов</div></div>{link}'
            f'<table><tr><th>Пакет</th><th class="n">Statements</th><th class="n">Покрытие</th><th></th></tr>'
            f'{"".join(rows)}</table>')
    write(site / "coverage" / "index.html", page("Coverage", body, "coverage", "../"))
    return pct


# ------------------------------------------------------------------- tests

def parse_tests(text: str, mod: str):
    pkgs: dict[str, dict] = collections.OrderedDict()
    output: dict[tuple, list[str]] = collections.defaultdict(list)
    failed: list[tuple] = []
    for line in text.splitlines():
        try:
            e = json.loads(line)
        except ValueError:
            continue
        pkg = e.get("Package")
        if not pkg:
            continue
        d = pkgs.setdefault(pkg, {"pass": 0, "fail": 0, "skip": 0, "elapsed": 0.0, "status": ""})
        test, action = e.get("Test"), e.get("Action")
        if test:
            if action in ("pass", "fail", "skip"):
                d[action] += 1
            if action == "fail":
                failed.append((pkg, test))
            if action == "output":
                output[(pkg, test)].append(e.get("Output", ""))
        else:
            if action in ("pass", "fail", "skip"):
                d["status"] = action
                d["elapsed"] = e.get("Elapsed", 0.0)
            if action == "output":
                output[(pkg, "")].append(e.get("Output", ""))
    return pkgs, output, failed


def short(pkg: str, mod: str) -> str:
    return pkg[len(mod) + 1:] if mod and pkg.startswith(mod + "/") else pkg


def build_tests(src: pathlib.Path, site: pathlib.Path, mod: str):
    text = read(src / "test" / "test.json")
    if text is None:
        write(site / "tests" / "index.html",
              page("Tests", '<div class="banner bad">Результаты тестов не получены.</div>', "tests", "../"))
        return None
    pkgs, output, failed = parse_tests(text, mod)
    tot = {"pass": 0, "fail": 0, "skip": 0}
    rows = []
    for pkg, d in pkgs.items():
        if d["pass"] + d["fail"] + d["skip"] == 0 and d["status"] != "fail":
            continue  # пакеты без тестов не показываем
        for k in tot:
            tot[k] += d[k]
        bad = d["status"] == "fail" or d["fail"] > 0
        rows.append(f'<tr><td class="p">{esc(short(pkg, mod))}</td>'
                    f'<td class="n ok">{d["pass"]}</td><td class="n {"bad" if bad else ""}">{d["fail"]}</td>'
                    f'<td class="n">{d["skip"]}</td><td class="n">{d["elapsed"]:.2f}s</td>'
                    f'<td class="{"bad" if bad else "ok"}">{"FAIL" if bad else "ok"}</td></tr>')
    broken = [p for p, d in pkgs.items() if d["status"] == "fail" and d["fail"] == 0]
    ok = tot["fail"] == 0 and not broken
    banner = ('<div class="banner ok">✅ Все тесты прошли</div>' if ok
              else '<div class="banner bad">❌ Есть упавшие тесты</div>')
    details = []
    for pkg, test in failed:
        out = "".join(output.get((pkg, test), []))
        details.append(f'<details open><summary>{esc(short(pkg, mod))} · {esc(test)}</summary>'
                       f'<pre>{esc(out)}</pre></details>')
    for pkg in broken:  # ошибка сборки пакета и т.п.
        out = "".join(output.get((pkg, ""), []))
        details.append(f'<details open><summary>{esc(short(pkg, mod))} · package</summary>'
                       f'<pre>{esc(out)}</pre></details>')
    body = (f'{banner}<div class="sub">unit + integration, -race. '
            f'Всего: {tot["pass"]} passed, {tot["fail"]} failed, {tot["skip"]} skipped</div>'
            f'<table><tr><th>Пакет</th><th class="n">Pass</th><th class="n">Fail</th>'
            f'<th class="n">Skip</th><th class="n">Время</th><th></th></tr>{"".join(rows)}</table>'
            f'{"".join(details)}')
    write(site / "tests" / "index.html", page("Tests", body, "tests", "../"))
    return {**tot, "ok": ok}


# -------------------------------------------------------------------- lint

def build_lint(src: pathlib.Path, site: pathlib.Path):
    text = read(src / "lint" / "lint.txt")
    status = (read(src / "lint" / "lint.status") or "").strip()
    if text is None:
        write(site / "lint" / "index.html",
              page("Lint", '<div class="banner bad">Вывод линтера не получен.</div>', "lint", "../"))
        return None
    issues = len(re.findall(r"^\S+\.go:\d+:\d+:", text, re.M))
    ok = status == "0"
    banner = ('<div class="banner ok">✅ Линтер прошёл — замечаний нет</div>' if ok
              else f'<div class="banner bad">❌ Замечаний: {issues}</div>')
    body = f'{banner}<pre>{esc(text.strip() or "замечаний нет")}</pre>'
    write(site / "lint" / "index.html", page("Lint", body, "lint", "../"))
    return {"ok": ok, "issues": issues}


# ------------------------------------------------------------------ readme

def badge_color(p: float) -> str:
    return ("brightgreen" if p >= 80 else "green" if p >= 60 else "yellow" if p >= 40
            else "orange" if p >= 20 else "red")


def update_readme(path: pathlib.Path, pct: float) -> None:
    text = path.read_text(encoding="utf-8")
    new = re.sub(r"img\.shields\.io/badge/Coverage-[0-9.]+%25-\w+",
                 f"img.shields.io/badge/Coverage-{pct}%25-{badge_color(pct)}", text)
    if new != text:
        path.write_text(new, encoding="utf-8")


# -------------------------------------------------------------------- main

def card(title: str, big: str, cls: str, sub: str, href: str) -> str:
    return (f'<a class="card" href="{href}"><h2>{title}</h2><div class="big {cls}">{big}</div>'
            f'<div class="sub">{sub}</div></a>')


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--reports", default="reports")
    ap.add_argument("--site", default="site")
    ap.add_argument("--readme", default="")
    args = ap.parse_args()
    src, site, mod = pathlib.Path(args.reports), pathlib.Path(args.site), module_name()
    site.mkdir(parents=True, exist_ok=True)

    cov = build_coverage(src, site, mod)
    tests = build_tests(src, site, mod)
    lint = build_lint(src, site)

    cards = [
        card("Coverage", f"{cov}%" if cov is not None else "—", pct_class(cov) if cov is not None else "bad",
             "бизнес-логика, internal/", "coverage/"),
        card("Tests", ("✓ " if tests["ok"] else "✗ ") + str(tests["pass"]) if tests else "—",
             "ok" if tests and tests["ok"] else "bad",
             f'{tests["fail"]} failed · {tests["skip"]} skipped' if tests else "нет данных", "tests/"),
        card("Lint", "✓" if lint and lint["ok"] else (f'{lint["issues"]}' if lint else "—"),
             "ok" if lint and lint["ok"] else "bad",
             "golangci-lint v2" if lint and lint["ok"] else "замечаний", "lint/"),
    ]
    write(site / "index.html", page("Отчёты о качестве кода", f'<div class="grid">{"".join(cards)}</div>',
                                    "index", ""))
    write(site / "summary.json", json.dumps({"coverage": cov, "tests": tests, "lint": lint}, indent=2))
    write(site / ".nojekyll", "")

    if args.readme and cov is not None:
        update_readme(pathlib.Path(args.readme), cov)
    print(f"coverage={cov} tests={tests} lint={lint}")


if __name__ == "__main__":
    main()
