#!/usr/bin/env python3
# /// script
# requires-python = ">=3.14"
# dependencies = [
#     "paddleocr>=3.6.0",
#     "requests>=2.34.2",
# ]
# ///
"""
Simple wrapper to run PaddleOCR on an image URL and print detected text as JSON array of strings.
Dependencies: paddleocr, requests
Example output: ["line1", "line2"]
"""
import sys
import json
import tempfile
import requests
import os

try:
    from paddleocr import PaddleOCR
except Exception as e:
    print(json.dumps({"error": f"missing dependency: {e}"}))
    sys.exit(2)


def extract_strings(obj, out):
    if isinstance(obj, str):
        out.append(obj)
    elif isinstance(obj, (list, tuple)):
        for v in obj:
            extract_strings(v, out)
    elif isinstance(obj, dict):
        for v in obj.values():
            extract_strings(v, out)


def main():
    if len(sys.argv) < 2:
        print(json.dumps({"error": "image url required"}))
        sys.exit(1)
    url = sys.argv[1]
    try:
        r = requests.get(url, timeout=30)
        r.raise_for_status()
    except Exception as e:
        print(json.dumps({"error": f"failed fetch image: {e}"}))
        sys.exit(2)

    tf = None
    try:
        tf = tempfile.NamedTemporaryFile(delete=False, suffix=os.path.splitext(url)[1] or ".jpg")
        tf.write(r.content)
        tf.close()

        ocr = PaddleOCR(use_angle_cls=True, lang='en')
        result = ocr.ocr(tf.name, cls=True)

        texts = []
        extract_strings(result, texts)
        # filter non-empty strings and dedupe ordering
        texts = [t for t in texts if isinstance(t, str) and t.strip()]

        print(json.dumps(texts))
    except Exception as e:
        print(json.dumps({"error": f"ocr error: {e}"}))
        sys.exit(2)
    finally:
        if tf is not None:
            try:
                os.unlink(tf.name)
            except Exception:
                pass


if __name__ == '__main__':
    main()
