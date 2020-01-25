# Copyright 2019 The Predictive Horizontal Pod Autoscaler Authors.
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

# CPA base image
FROM registry.access.redhat.com/ubi8/ubi


ENV USER_UID=1001 \
    USER_NAME=custom-pod-autoscaler-operator

# Install sqlite and container deps
RUN dnf install sqlite wget -y

# Install CPA
RUN wget -qO- https://github.com/jthomperoo/custom-pod-autoscaler/releases/download/v0.10.0/custom-pod-autoscaler.tar.gz | tar xvz && mv dist cpa

WORKDIR /app

ENV configPath /app/config.yaml
# Add in entry point script
COPY build/entrypoint.sh /
# Add in config and horizontal-pod-autoscaler binary
COPY config.yaml dist/* ./
# Add in db migrations
COPY sql ./sql

RUN mkdir -p ${HOME} && \
    chown ${USER_UID}:0 ${HOME} && \
    chmod ug+rwx $HOME && \
    chmod g+rw /etc/passwd

RUN mkdir /store && \
    chown ${USER_UID}:0 /store

USER ${USER_UID}

CMD [ "/entrypoint.sh" ]