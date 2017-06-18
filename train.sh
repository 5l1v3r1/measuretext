#!/bin/bash

if [ ! -f out_net ]; then
  echo 'Creating network...'
  trap 'rm -r tmp' EXIT
  mkdir tmp
  for direc in forward backward
  do
    neurocli new -in rnn_block.txt -out tmp/$direc
  done
  neurocli new -in mixer.txt -out tmp/mixer
  neurocli bidir -forward tmp/forward -backward tmp/backward \
    -mixer tmp/mixer -out out_net
  rm -r tmp
  trap '' EXIT
fi

echo 'Training...'
neurocli train -adam default -samples data/data.txt -net out_net \
  -cost mse -batch 128

echo 'Validating...'

neurocli train -step 0 -samples data/val.txt -net out_net \
  -cost mse -batch 100 -stopsamples 399
