#!/usr/bin/env python3
"""Fix uno external test files after migration from internal package."""
import re
from pathlib import Path

UNO_DIR = Path(__file__).resolve().parent / "uno"

COLOR_VALUE_EVENT = [
    ("ColorRed", "uno.ColorRed"),
    ("ColorBlue", "uno.ColorBlue"),
    ("ColorGreen", "uno.ColorGreen"),
    ("ColorYellow", "uno.ColorYellow"),
    ("ColorWild", "uno.ColorWild"),
    ("ValueReverse", "uno.ValueReverse"),
    ("ValueDraw2", "uno.ValueDraw2"),
    ("ValueSkip", "uno.ValueSkip"),
    ("ValueWild4", "uno.ValueWild4"),
    ("EventDraw", "uno.EventDraw"),
    ("EventPlay", "uno.EventPlay"),
    ("EventPlayerOut", "uno.EventPlayerOut"),
]

for path in UNO_DIR.glob("*_test.go"):
    text = path.read_text()
    text = text.replace("uno.Card:", "Card:")
    text = text.replace(".uno.Card", ".Card")
    text = text.replace("filterEventsForSeat", "uno.FilterEventsForSeat")
    text = re.sub(r"(?<![.\w])RunAITurns\b", "uno.RunAITurns", text)
    for old, new in COLOR_VALUE_EVENT:
        text = re.sub(rf"(?<![.\w]){re.escape(old)}\b", new, text)
    path.write_text(text)
