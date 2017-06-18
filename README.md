# measuretext

This is a small experiment to see how well a neural network can predict the width of rendered text. Measuring text is a fairly common operation in UI development. I think it would be cool to bake a font's rendering dynamics into a neural network.

This uses [neurocli](https://github.com/unixpickle/neurocli) for training a network. It uses Google Chrome to measure text when producing data.
