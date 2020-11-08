import pandas as pd
from pmdarima.arima import auto_arima
from json import JSONEncoder
import pickle
import numpy
import json


from flask import Flask, request, jsonify
app = Flask(__name__)
MODEL_LOCATION='/tmp/arima.pkl'


class NumpyArrayEncoder(JSONEncoder):
    def default(self, obj):
        if isinstance(obj, numpy.ndarray):
            return obj.tolist()
        return JSONEncoder.default(self, obj)

# Example post command for predict: curl -XPOST  localhost:5000/api/train -d '{"index": {"1": 5, "2": 2, "3": 7, "4": 6}, "value": {"1": 8, "2": 4, "3": 1, "4": 9}}'  -H 'content-type: application/json'
# one can use test.py to generate the data set
@app.route('/api/train', methods=['POST'])
def train():
    content = request.json
    data = pd.DataFrame.from_dict(content,orient='index').T.set_index('index')   
    print(data)
    stepwise_model = auto_arima(data, start_p=1, start_q=1,
                           max_p=3, max_q=3, m=12,
                           start_P=0, seasonal=True,
                           d=1, D=1, trace=True,
                           error_action='ignore',  
                           suppress_warnings=True, 
                           stepwise=True)
    print(stepwise_model.aic())
    print(stepwise_model.summary())

    # Serialize with Pickle
    with open('/tmp/arima.pkl', 'wb') as pkl:
        pickle.dump(stepwise_model, pkl)
    # print(p)
    return jsonify('{}')

def load_pred(location):
    f  = open(location, 'rb') 
    return pickle.load(f)

def update_model(location, data):
    model = load_pred(location)
    model.update(data)
    with open(location, 'wb') as pkl:
        pickle.dump(model, pkl)
    return model

# Example post command for predict: curl -XPOST  localhost:5000/api/predict -d '{"index": {"1": 5, "2": 2, "3": 7, "4": 6}, "value": {"1": 8, "2": 4, "3": 1, "4": 9}}'  -H 'content-type: application/json'
@app.route('/api/predict', methods=['POST'])
def predict():
    content = request.json
    data = pd.DataFrame.from_dict(content,orient='index').T.set_index('index') 
    model = update_model(MODEL_LOCATION, data)
    prediction, new_conf_int = model.predict(n_periods=10, return_conf_int=True)
    return json.dumps(prediction,cls=NumpyArrayEncoder), 



if __name__ == '__main__':
    app.run(host= '0.0.0.0',debug=True)