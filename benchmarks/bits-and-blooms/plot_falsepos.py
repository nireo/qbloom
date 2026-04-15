#!/usr/bin/env python3

import csv
import sys

import matplotlib.pyplot as plt
from matplotlib import colormaps
from matplotlib.ticker import FuncFormatter


plt.rcParams["font.size"] = 18

PROBE_BUDGET = 2_000_000
TRIALS = 16
DETECTION_FLOOR = 100.0 / (PROBE_BUDGET * TRIALS)


def read_series(path: str):
    rows = []
    with open(path, newline="", encoding="utf-8") as handle:
        reader = csv.reader(handle)
        for row in reader:
            if len(row) != 4:
                continue
            rows.append(tuple(float(value) for value in row))
    return rows


def format_y(value, _pos):
    if value >= 1:
        return f"{int(value)}"
    return f"{value:.8f}".rstrip("0").rstrip(".")


def main() -> int:
    if len(sys.argv) != 4:
        print(
            "usage: plot_falsepos.py <qbloom.csv> <bits-and-blooms.csv> <output.png>",
            file=sys.stderr,
        )
        return 1

    qbloom_rows = read_series(sys.argv[1])
    bits_rows = read_series(sys.argv[2])
    if not qbloom_rows or not bits_rows:
        print("input csv is empty", file=sys.stderr)
        return 1

    cm = [colormaps["Dark2"](i / 8) for i in range(8)]
    fig, ax = plt.subplots(1, 1, figsize=(10, 6.2))

    for name, rows, color in [
        ("qbloom", qbloom_rows, cm[0]),
        ("bits-and-blooms", bits_rows, cm[1]),
    ]:
        x = []
        avg = []
        for row in rows:
            x.append(row[0])
            value = row[1] * 100.0
            if value <= 0:
                value = DETECTION_FLOOR
            avg.append(value)
        ax.plot(x, avg, color=color, label=name, linewidth=2.2)

    ax.set_yscale("log")
    ax.set_title("Bloom Filter False Positives (Lower is Better)")
    ax.set_xlabel("Items Per Bit")
    ax.set_ylabel("False Positive %")
    ax.set_xlim(left=0)
    ax.set_xlim(right=0.125)
    ax.set_ylim(bottom=0.00000001, top=100)
    ax.yaxis.set_major_formatter(FuncFormatter(format_y))
    ax.grid(True, which="major", alpha=0.35)
    ax.legend(loc="lower right")
    fig.tight_layout()
    fig.savefig(sys.argv[3], dpi=180)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
