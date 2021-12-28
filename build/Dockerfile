# Copyright 2021 The Predictive Horizontal Pod Autoscaler Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Build stage, must be done in Docker for cross compatiblity because CGO is required for go-sqlite3
FROM golang:1.16 AS build

WORKDIR /app

COPY . ./

RUN make

# CPA base image
FROM custompodautoscaler/python-3-8:v2.3.0

ENV USER_UID=1001 \
    USER_NAME=phpa

# Install sqlite and container deps
RUN apt-get -y update && apt-get install -y sqlite3 libsqlite3-dev

WORKDIR /app

# Install any python requirements
COPY algorithms/requirements.txt ./algorithms/requirements.txt
RUN pip install -r ./algorithms/requirements.txt

ENV configPath /app/config.yaml
# Add in entry point script
COPY build/entrypoint.sh /
# Add in config and horizontal-pod-autoscaler binary
COPY config.yaml ./
# Add in db migrations
COPY sql ./sql
# Add in Python aglorithms
COPY algorithms ./algorithms
# Add in built binary
COPY --from=build /app/dist/* ./

RUN mkdir -p ${HOME} && \
    chown ${USER_UID}:0 ${HOME} && \
    chmod ug+rwx $HOME && \
    chmod g+rw /etc/passwd

RUN mkdir /store && \
    chown ${USER_UID}:0 /store

RUN useradd -u ${USER_UID} ${USER_NAME}

USER ${USER_UID}

CMD [ "/entrypoint.sh" ]
