#!/usr/bin/python
# -*- coding: utf-8 -*-
"""
Sarima algorithm for prediction.
# Example post command for predict:
# curl -XPOST  localhost:5000/api/predict -d '
# {"index": {"1": 5, "2": 2, "3": 7, "4": 6},
# "value": {"1": 8, "2": 4, "3": 1, "4": 9}}'
# -H 'content-type: application/json'
"""
from json import JSONEncoder
import json
import pickle
import numpy
import pandas as pd

from flask import Flask, request, jsonify
from pmdarima.arima import auto_arima


app = Flask(__name__)
MODEL_LOCATION = '/tmp/arima.pkl'


class NumpyArrayEncoder(JSONEncoder):
    '''
    NumpyArraryEncoder class
    '''

    def default(self, o):
        '''
        obj: ndarray to json
        '''
        if isinstance(o, numpy.ndarray):
            return o.tolist()
        return JSONEncoder.default(self, o)


@app.route('/api/train', methods=['POST'])
def train():
    '''
    Post method to return default json object with 200OK
    '''
    content = request.json
    data = pd.DataFrame.from_dict(content, orient='index'
                                  ).T.set_index('index')
    stepwise_model = auto_arima(
        data,
        start_p=1,
        start_q=1,
        max_p=3,
        max_q=3,
        m=12,
        start_P=0,
        seasonal=True,
        d=1,
        D=1,
        trace=True,
        error_action='ignore',
        suppress_warnings=True,
        stepwise=True,
        )
    print(stepwise_model.aic())
    print(stepwise_model.summary())

    # Serialize with Pickle

    with open('/tmp/arima.pkl', 'wb') as pkl:
        pickle.dump(stepwise_model, pkl)

    return jsonify('{}')


def load_pred(location):
    '''
    Load prediction from location
    location: default path to load prediction from
    Returns: Picket object
    '''
    f_d = open(location, 'rb')
    return pickle.load(f_d)


def update_model(location, data):
    '''
    Update model with latest data
    location: Location of model to load from
    data: new data to update the model with
    Returns: pickled model
    '''
    model = load_pred(location)
    model.update(data)
    with open(location, 'wb') as pkl:
        pickle.dump(model, pkl)
    return model


@app.route('/api/predict', methods=['POST'])
def predict():
    '''
    Load prediction and update them with existing model
    Returns: Jsonify data
    '''
    content = request.json
    data = pd.DataFrame.from_dict(content, orient='index'
                                  ).T.set_index('index')
    model = update_model(MODEL_LOCATION, data)
    (prediction, new_conf_int) = model.predict(n_periods=10, return_conf_int=True)
    print(new_conf_int)
    return json.dumps(prediction, cls=NumpyArrayEncoder)

if __name__ == '__main__':
    app.run(host='0.0.0.0', debug=True)
