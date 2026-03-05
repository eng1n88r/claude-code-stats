"""Tests for extract_stats.py core functions.

Tests the pure logic functions without requiring ~/.claude/ data.
"""
import json
import math
import os
import sys
import unittest
from pathlib import Path

# Add parent dir so we can import extract_stats functions
sys.path.insert(0, str(Path(__file__).parent))

# We need to mock CONFIG before importing extract_stats since it loads at module level
# Patch the config loading to use test fixtures
TESTDATA = Path(__file__).parent.parent / "testdata"
os.environ["CLAUDE_STATS_TEST"] = "1"


class TestPricing(unittest.TestCase):
    """Test pricing and cost calculation functions."""

    def setUp(self):
        # Import fresh — module-level globals already loaded
        import extract_stats
        self.mod = extract_stats

    def test_calc_cost_opus(self):
        """Test cost calculation for Opus 4.5 model."""
        usage = {
            "input_tokens": 1000,
            "output_tokens": 500,
            "cache_read_input_tokens": 200,
            "cache_creation_input_tokens": 100,
        }
        cost = self.mod.calc_cost("claude-opus-4-5-20251101", usage)
        # 1000*5/1M + 500*25/1M + 200*0.5/1M + 100*6.25/1M
        # = 0.005 + 0.0125 + 0.0001 + 0.000625 = 0.018225
        self.assertAlmostEqual(cost, 0.018225, places=6)

    def test_calc_cost_haiku(self):
        """Test cost calculation for Haiku 4.5 model."""
        usage = {
            "input_tokens": 500,
            "output_tokens": 100,
            "cache_read_input_tokens": 0,
            "cache_creation_input_tokens": 0,
        }
        cost = self.mod.calc_cost("claude-haiku-4-5-20251001", usage)
        # 500*1/1M + 100*5/1M = 0.0005 + 0.0005 = 0.001
        self.assertAlmostEqual(cost, 0.001, places=6)

    def test_calc_cost_zero_tokens(self):
        cost = self.mod.calc_cost("claude-opus-4-6", {})
        self.assertEqual(cost, 0.0)

    def test_calc_cost_unknown_model(self):
        """Unknown model uses default (Opus-level) pricing."""
        usage = {"input_tokens": 1_000_000, "output_tokens": 0}
        cost = self.mod.calc_cost("claude-mystery", usage)
        self.assertAlmostEqual(cost, 5.00, places=2)

    def test_calc_cost_large_tokens(self):
        usage = {
            "input_tokens": 100_000,
            "output_tokens": 50_000,
            "cache_read_input_tokens": 500_000,
            "cache_creation_input_tokens": 10_000,
        }
        cost = self.mod.calc_cost("claude-opus-4-6", usage)
        # 100000*5/1M + 50000*25/1M + 500000*0.5/1M + 10000*6.25/1M
        # = 0.5 + 1.25 + 0.25 + 0.0625 = 2.0625
        self.assertAlmostEqual(cost, 2.0625, places=4)

    def test_get_model_display(self):
        tests = {
            "claude-opus-4-6": "Opus 4.6",
            "claude-opus-4-5-20251101": "Opus 4.5",
            "claude-sonnet-4-5-20250929": "Sonnet 4.5",
            "claude-haiku-4-5-20251001": "Haiku 4.5",
            "claude-future-model": "Unknown",
        }
        for model_id, expected in tests.items():
            with self.subTest(model_id=model_id):
                self.assertEqual(self.mod.get_model_display(model_id), expected)


class TestProjectDisplayName(unittest.TestCase):
    def setUp(self):
        import extract_stats
        self.fn = extract_stats.project_display_name

    def test_unix_path(self):
        self.assertEqual(self.fn("/home/user/projects/myapp"), "projects/myapp")

    def test_windows_path(self):
        self.assertEqual(self.fn("C:\\Users\\me\\work\\dashboard"), "work/dashboard")

    def test_single_component(self):
        self.assertEqual(self.fn("/single"), "single")

    def test_empty(self):
        self.assertEqual(self.fn(""), "Unknown")

    def test_none(self):
        self.assertEqual(self.fn(None), "Unknown")

    def test_deep_path(self):
        self.assertEqual(self.fn("/a/b/c/d/e"), "d/e")

    def test_trailing_slash(self):
        self.assertEqual(self.fn("/trailing/slash/"), "trailing/slash")


class TestCLIHelpers(unittest.TestCase):
    """Test cli.py helper functions."""

    def setUp(self):
        import cli
        self.cli = cli

    def test_fmt_cost(self):
        self.assertEqual(self.cli.fmt_cost(0), "$0.00")
        self.assertEqual(self.cli.fmt_cost(18.59), "$18.59")
        self.assertEqual(self.cli.fmt_cost(1234.5), "$1,234.50")
        self.assertEqual(self.cli.fmt_cost(None), "N/A")

    def test_fmt_tokens(self):
        self.assertEqual(self.cli.fmt_tokens(0), "0")
        self.assertEqual(self.cli.fmt_tokens(500), "500")
        self.assertEqual(self.cli.fmt_tokens(1000), "1.0K")
        self.assertEqual(self.cli.fmt_tokens(6300), "6.3K")
        self.assertEqual(self.cli.fmt_tokens(1_000_000), "1.0M")
        self.assertEqual(self.cli.fmt_tokens(1_200_000), "1.2M")
        self.assertEqual(self.cli.fmt_tokens(None), "0")

    def test_make_bar(self):
        bar = self.cli.make_bar(50, 100, width=10)
        self.assertEqual(len(bar), 10)
        self.assertEqual(bar.count("\u2588"), 5)  # filled blocks
        self.assertEqual(bar.count("\u2591"), 5)  # empty blocks

    def test_make_bar_zero(self):
        self.assertEqual(self.cli.make_bar(0, 0), "")

    def test_make_bar_full(self):
        bar = self.cli.make_bar(100, 100, width=10)
        self.assertEqual(bar, "\u2588" * 10)


class TestPricingParity(unittest.TestCase):
    """Verify Python and Go produce identical cost calculations.

    These test cases use the same inputs as the Go TestCalcCost tests.
    If both pass, the implementations are equivalent.
    """

    def setUp(self):
        import extract_stats
        self.calc_cost = extract_stats.calc_cost

    def test_parity_opus_basic(self):
        """Same as Go TestCalcCost/opus_4.5_basic."""
        usage = {
            "input_tokens": 1000,
            "output_tokens": 500,
            "cache_read_input_tokens": 200,
            "cache_creation_input_tokens": 100,
        }
        cost = self.calc_cost("claude-opus-4-5-20251101", usage)
        self.assertAlmostEqual(cost, 0.018225, places=6)

    def test_parity_haiku_basic(self):
        """Same as Go TestCalcCost/haiku_basic."""
        usage = {
            "input_tokens": 500,
            "output_tokens": 100,
            "cache_read_input_tokens": 0,
            "cache_creation_input_tokens": 0,
        }
        cost = self.calc_cost("claude-haiku-4-5-20251001", usage)
        self.assertAlmostEqual(cost, 0.001, places=6)

    def test_parity_zero(self):
        """Same as Go TestCalcCost/zero_tokens."""
        cost = self.calc_cost("claude-opus-4-6", {})
        self.assertEqual(cost, 0.0)

    def test_parity_unknown(self):
        """Same as Go TestCalcCost/unknown_model_uses_default_pricing."""
        usage = {"input_tokens": 1_000_000, "output_tokens": 0}
        cost = self.calc_cost("claude-mystery", usage)
        self.assertAlmostEqual(cost, 5.00, places=2)

    def test_parity_large(self):
        """Same as Go TestCalcCost/large_token_counts."""
        usage = {
            "input_tokens": 100_000,
            "output_tokens": 50_000,
            "cache_read_input_tokens": 500_000,
            "cache_creation_input_tokens": 10_000,
        }
        cost = self.calc_cost("claude-opus-4-6", usage)
        self.assertAlmostEqual(cost, 2.0625, places=4)


class TestProjectDisplayNameParity(unittest.TestCase):
    """Same test cases as Go TestProjectDisplayName."""

    def setUp(self):
        import extract_stats
        self.fn = extract_stats.project_display_name

    def test_parity_unix(self):
        self.assertEqual(self.fn("/home/user/projects/myapp"), "projects/myapp")

    def test_parity_windows(self):
        self.assertEqual(self.fn("C:\\Users\\me\\work\\dashboard"), "work/dashboard")

    def test_parity_single(self):
        self.assertEqual(self.fn("/single"), "single")

    def test_parity_empty(self):
        self.assertEqual(self.fn(""), "Unknown")

    def test_parity_deep(self):
        self.assertEqual(self.fn("/a/b/c/d/e"), "d/e")

    def test_parity_trailing_slash(self):
        self.assertEqual(self.fn("/trailing/slash/"), "trailing/slash")


if __name__ == "__main__":
    unittest.main()
