FROM jupyter/base-notebook:python-3.10

# for offline builds, add ability for custom PIP repositories
ARG PIP_INDEX
ARG PIP_INDEX_URL
ARG PIP_TRUSTED_HOST

# RUN pip3 install mythic==0.1.2

WORKDIR /projects

#CMD start.sh jupyter lab --ServerApp.open_browser=false --IdentityProvider.token='' --ServerApp.base_url="/jupyter" --ServerApp.default_url="/jupyter"

ENV JUPYTERHUB_SERVICE_PREFIX "/jupyter/"

ENV JUPYTER_ENABLE_LAB=yes

USER root
RUN apt-get update \
    && apt-get -y upgrade \
    && apt-get install -y --no-install-recommends build-essential git\
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*
USER 1000
RUN pip3 install -U --use-deprecated=legacy-resolver --no-cache-dir \
    torch \
    numpy \
    matplotlib \
    adversarial-robustness-toolbox \
    markupsafe==2.0.1 \
    scikit-learn \
    pandas \
    transformers \
    datasets \
    optuna \
    ipywidgets \
    myst-parser \
    jupyterlab_myst \ 
    diffusers\
    torchvision\
    mythic==0.1.2
