# Inverting a 3x2 Affine Transformation Matrix

In 2-D geometry, a coordinate pair `(x, y)` can be thought of as a 2x1 matrix
(a vector). [Affine
transformations](https://en.wikipedia.org/wiki/Affine_transformation)
(including rotations, scales, translations and combinations of those) can be
represented by a 3x2 matrix `F`:

    [ Fa   Fb   Fc ]
    [ Fd   Fe   Ff ]

Given a starting point `S = (Sx, Sy)`, the transformed point `T = (Tx, Ty)` is
given by two scalar equations:

    Tx = (Fa * Sx) + (Fb * Sy) + Fc
    Ty = (Fd * Sx) + (Fe * Sy) + Ff

There is an equivalent
[homogenous](https://en.wikipedia.org/wiki/Homogeneous_coordinates) form, where
the 3x2 matrix `F` is augmented with a `[0 0 1]` bottom row and the 2x1 vectors
`S` and `T` are augmented with a `[1]` bottom element. Transformation becomes
equivalent to matrix multiplication, as these three scalar equations:

    Tx = (Fa * Sx) + (Fb * Sy) + (Fc * 1)
    Ty = (Fd * Sx) + (Fe * Sy) + (Ff * 1)
    1  = ( 0 * Sx) + ( 0 * Sy) + ( 1 * 1)

Are equivalent to the one matrix equation `T = F * S`:

    [ Tx ]     [ Fa   Fb   Fc ]     [ Sx ]
    [ Ty ]  =  [ Fd   Fe   Ff ]  *  [ Sy ]
    [  1 ]     [  0    0    1 ]     [  1 ]

That (augmented) forward matrix `F` maps from `S` to `T`. As long as the
determinant `FΔ` (see below) is non-zero, the inverse transformation, mapping
from `T` to `S`, exists. If so, it is also an affine transformation and
representable by a 3x2 (or augmented 3x3) backward matrix `B` such that
`S = B * T`:

    [ Sx ]     [ Ba   Bb   Bc ]     [ Tx ]
    [ Sy ]  =  [ Bd   Be   Bf ]  *  [ Ty ]
    [  1 ]     [  0    0    1 ]     [  1 ]

`B` is the inverse of `F`, and there is a formula for the inverse of a 3x3
matrix, but it's somewhat messy in general. The formula can be simplified when
the bottom row is `[0 0 1]`. I tried for a few minutes to find this simplified
formula online, but I failed. It was easy to simplify the general formula
myself, and since I did the algebra, I thought I might as well write it out as
this very blog post.

Specifically, the inverse `B` of the 3x2 matrix `F` (given at the top) is:

    [ Ba   Bb   Bc ]
    [ Bd   Be   Bf ]

Where:

    FΔ = (Fa * Fe) - (Fb * Fd)
    Ba = +Fe / FΔ
    Bb = -Fb / FΔ
    Bc = ((Fb * Ff) - (Fe * Fc)) / FΔ
    Bd = -Fd / FΔ
    Be = +Fa / FΔ
    Bf = ((Fd * Fc) - (Fa * Ff)) / FΔ

As long as the [determinant](https://en.wikipedia.org/wiki/Determinant) `FΔ` is
non-zero (so that the `F` matrix is invertible), it's straightforward to
confirm that `F * B = I`, the identity matrix (and likewise `B * F = I`). For
example:

    [ Fa   Fb   Fc ]     [ +Fe/FΔ   -Fb/FΔ   ((Fb*Ff)-(Fe*Fc))/FΔ ]     [ 1 0 0 ]
    [ Fd   Fe   Ff ]  *  [ -Fd/FΔ   +Fa/FΔ   ((Fd*Fc)-(Fa*Ff))/FΔ ]  =  [ 0 1 0 ]
    [  0    0    1 ]     [      0        0                      1 ]     [ 0 0 1 ]


---

Published: 2021-12-30
