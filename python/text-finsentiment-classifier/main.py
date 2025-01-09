import os 

from flask import Flask, request, jsonify
from pydantic import BaseModel, ValidationError
from typing import List

from transformers import BertTokenizer, BertForSequenceClassification
from transformers import pipeline

app = Flask(__name__)

# Load the classifier outside the request handler for efficiency
try:
    model = BertForSequenceClassification.from_pretrained("ahmedrachid/FinancialBERT-Sentiment-Analysis",num_labels=3)
    tokenizer = BertTokenizer.from_pretrained("ahmedrachid/FinancialBERT-Sentiment-Analysis")

    classifier = pipeline("sentiment-analysis", model=model, tokenizer=tokenizer)
except Exception as e:
    print(f"Error loading classifier: {e}")
    exit(1) #Exit if the classifier fails to load

class TextItem(BaseModel):
    text: str

class RequestData(BaseModel):
    items: List[TextItem]

@app.route("/classify", methods=["POST"])
def classify_text():
    try:
        data = request.get_json()
        if data is None:
            return jsonify({"error": "No JSON data provided"}), 400

        request_data = RequestData(**data)
        try:
            to_classify = [i.text for i in request_data.items]
            results = classifier(to_classify)
        except Exception as e:
            print(f"Error classifying text: {e}")
            results = [{"error": f"Error classifying text: {e}"}]

        return jsonify(results), 200

    except ValidationError as e:
        return jsonify({"error": e.errors()}), 400
    except Exception as e:
        print(f"Unexpected error: {e}")
        return jsonify({"error": "An unexpected error occurred"}), 500


if __name__ == "__main__":
    app.run(debug=True, host="0.0.0.0", port=3001)
