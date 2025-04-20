import unittest
import tempfile
import os
import shutil
from logic.storage import load, save, list_saved_games

class TestStorage(unittest.TestCase):
    def setUp(self):
        self.test_dir = tempfile.mkdtemp()
        self.saves_dir = os.path.join(self.test_dir, "data", "saves")
        os.makedirs(self.saves_dir)

    def tearDown(self):
        shutil.rmtree(self.test_dir)

    def test_storage_interface(self):
        test_data = {"test": "data"}
        save_path = os.path.join(self.saves_dir, "test.json")
        
        save(test_data, save_path)
        self.assertTrue(os.path.exists(save_path))
        
        loaded_data = load(save_path)
        self.assertEqual(loaded_data, test_data)
        
        saves = list_saved_games(self.saves_dir)
        self.assertIn("test.json", saves)

if __name__ == '__main__':
    unittest.main() 
