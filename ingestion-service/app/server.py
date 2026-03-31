import logging

from app.config import load_config
from app.service import serve


if __name__ == "__main__":
    cfg = load_config()
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
    serve(cfg)
