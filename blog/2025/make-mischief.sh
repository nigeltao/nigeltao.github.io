#!/bin/bash

lilypond mischief.ly

pdftoppm mischief.pdf x
for i in {1..3}; do
	convert x-$i.ppm mischief.$i.png;
	pngcrush -ow -brute -rem alla mischief.$i.png
	rm x-$i.ppm
done

fluidsynth -F x0.wav mischief.midi
sox x0.wav x1.wav gain -2 reverb 50 50 100
lame x1.wav mischief.mp3
rm x?.wav
