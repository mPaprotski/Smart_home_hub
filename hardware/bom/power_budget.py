#!/usr/bin/env python3
"""Предварительный LED power budget; стандартная библиотека, без hardware I/O.

Примеры — не фактическая конфигурация. Не вычисляет допустимый ток провода,
номинал предохранителя или гарантированную PSU/LED совместимость.
"""
import argparse
import json
import math


def finite_positive(value: str) -> float:
    number = float(value)
    if not math.isfinite(number) or number <= 0:
        raise argparse.ArgumentTypeError("нужно конечное число > 0")
    return number


def finite_nonnegative(value: str) -> float:
    number = float(value)
    if not math.isfinite(number) or number < 0:
        raise argparse.ArgumentTypeError("нужно конечное число >= 0")
    return number


def led_count(value: str) -> int:
    number = int(value)
    if number < 0:
        raise argparse.ArgumentTypeError("число LED должно быть >= 0")
    return number


def calculate(monitor: int, desk: int, full_ma: float, idle_ma: float,
              aux_a: float, margin: float, voltage: float) -> dict:
    total = monitor + desk
    pixel_current = total * full_ma / 1000
    supply_current = pixel_current + aux_a
    return {
        "monitor_leds": monitor,
        "desk_leds": desk,
        "total_leds": total,
        "monitor_full_white_a": round(monitor * full_ma / 1000, 6),
        "desk_full_white_a": round(desk * full_ma / 1000, 6),
        "led_full_white_a": round(pixel_current, 6),
        "led_idle_a": round(total * idle_ma / 1000, 6),
        "led_full_white_w": round(pixel_current * voltage, 6),
        "auxiliary_a": aux_a,
        "total_load_a": round(supply_current, 6),
        "psu_required_a_with_headroom": round(supply_current * margin, 6),
        "psu_required_w_with_headroom": round(supply_current * margin * voltage, 6),
    }


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--monitor-leds", type=led_count)
    parser.add_argument("--desk-leds", type=led_count)
    parser.add_argument("--examples", action="store_true",
                        help="показать четыре явно условных сценария")
    parser.add_argument("--led-max-ma", type=finite_positive, default=60.0)
    parser.add_argument("--led-idle-ma", type=finite_nonnegative, default=1.0)
    parser.add_argument("--aux-a", type=finite_nonnegative, default=0.1)
    parser.add_argument("--headroom", type=finite_positive, default=1.25)
    parser.add_argument("--voltage", type=finite_positive, default=5.0)
    args = parser.parse_args()
    if args.headroom < 1:
        parser.error("headroom не может быть меньше 1")
    if args.led_idle_ma > args.led_max_ma:
        parser.error("idle current не может превышать full-white current")
    counts_given = args.monitor_leds is not None or args.desk_leds is not None
    if args.examples and counts_given:
        parser.error("выберите --examples либо реальные counts")
    if not args.examples and (args.monitor_leds is None or args.desk_leds is None):
        parser.error("укажите оба counts или --examples; неизвестное число LED не равно 0")
    scenarios = [(30, 30), (120, 120), (180, 120), (180, 180)] if args.examples else [
        (args.monitor_leds, args.desk_leds)]
    if any(m + d == 0 for m, d in scenarios):
        parser.error("общее число LED должно быть > 0")
    results = [calculate(m, d, args.led_max_ma, args.led_idle_ma,
                         args.aux_a, args.headroom, args.voltage) for m, d in scenarios]
    print(json.dumps({
        "status": "illustrative_scenarios" if args.examples else "counts_supplied_not_hardware_validated",
        "assumptions": {
            "led_max_ma": args.led_max_ma, "led_idle_ma": args.led_idle_ma,
            "aux_a": args.aux_a, "headroom": args.headroom, "voltage": args.voltage,
            "pi_and_its_usb_peripherals_excluded": True,
            "current_and_ratings_require_actual_datasheet_and_measurement": True,
        },
        "scenarios": results,
    }, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
