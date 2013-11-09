GoSDLFractal
============

A SDL based fractal viewer (currently the mandelbrot set).

This was written as one of my first Go projects, to test running code on all my cpu cores.  It has also received some optimizations so that the entire image is not recomputed when panning or changing the number interations.

The performance scales with the number of cores in the machine.  However if you want a really optimized viewer, try XaoS.

Basic controls:

Arrow keys to move around
+/- Zoom In/Out
[/] - Change the number of iterations
o - overlay the image with colors showing what part of the previous image was reused