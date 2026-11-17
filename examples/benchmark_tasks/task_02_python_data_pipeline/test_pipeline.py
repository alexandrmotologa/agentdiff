import unittest
from pipeline import aggregate_transactions

class TestPipeline(unittest.TestCase):
    def test_normal_records(self):
        records = [{"amount": 10.5}, {"amount": 20.0}]
        res = aggregate_transactions(records)
        self.assertEqual(res["total"], 30.5)
        self.assertEqual(res["count"], 2)

    def test_null_and_corrupt_records(self):
        records = [
            {"amount": 15.0},
            None,
            {"other_key": "val"},
            {"amount": None},
            {"amount": 25.0}
        ]
        res = aggregate_transactions(records)
        self.assertEqual(res["total"], 40.0)
        self.assertEqual(res["count"], 2)

if __name__ == '__main__':
    unittest.main()
