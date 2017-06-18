#!/bin/bash

echo 'Validating...'

neurocli train \
  -step 0 \
  -samples data/val.txt \
  -net out_net \
  -cost mse \
  -batch 100 \
  -stopsamples 399
