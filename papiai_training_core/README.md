# Sovereign Intelligence Core - PapiAi Training Matrix

This directory contains the proprietary neural training and inference pipeline for PapiAi. It automatically ingests open-source datasets and foundation models from the internet, completely rebrands them into the **PapiAi Sovereign Intelligence** identity, and compiles a local neural engine.

## Prerequisites

Ensure you have a working Python environment (Python 3.10+ recommended).
Install the core requirements:
```bash
pip install -r requirements.txt
```

## 1. Train the PapiAi Engine
To aggregate the dataset, format it with the PapiAi identity, and compile your own local weights:
```bash
python3 train_papiai.py
```
*Note: This will download a lightweight foundation model, inject your custom identity, and save the compiled weights to `./papiai_neural_weights`. Depending on your hardware, this may take some time. It is configured to run on CPU if no GPU is available.*

## 2. Launch the Neural Inference
Once trained (or if you just want to test the base engine), run the inference script to chat directly with PapiAi in your terminal:
```bash
python3 run_papiai.py
```

## Features
- **Zero Telemetry**: All standard telemetry (e.g., WandB) has been stripped out to maintain 100% data sovereignty.
- **Identity Injection**: Automatically rewrites dataset conversational structures to establish the "Sovereign Intelligence Core" persona.
- **Lightweight Architecture**: Uses a highly efficient 0.5B parameter base to ensure it can compile locally without massive server farms.
