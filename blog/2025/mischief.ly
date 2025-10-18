\version "2.24.1"
\include "english.ly"
\pointAndClickOff

\header {
  title = "Mischief"
  composer = "by Nigel Tao"
  tagline = #f

  creationDate = ""
  modDate = ""
  pdfCreationDate = ""
  pdfModDate = ""
}

the_upper = {
  \tempo 4 = 160

  \set Staff.midiInstrument = "harpsichord"
  \clef treble
  \key a \minor
  \time 4/4

  a'4. e''8            ds''  e'' f''  e''
  a'4. e''8            f''4      e''
  a'4. e''8            ds''  e'' ds'' e''
  a'4  d''             c''4      b'8  a'

  g'4. d''8            cs''  d'' ef'' d''
  g'4. d''8            ef''4     d''
  g'4. d''8            cs''  d'' c''  d''
  bf'8 d'' a' d''      g'    d'' f'   d''

  ef'4. bf'8           a'    bf' a'   bf'
  ef'4. bf'8           c''4      bf'
  d'4.  bf'8           a'    bf' a'   bf'
  c'8  bf' a' g'       a'    fs' g'   gs'

  \bar "||" %% Bar  13

  a'4. e''8            ds''  e'' f''  e''
  a'4. e''8            f''4      e''
  a'4. e''8            ds''  e'' ds'' e''
  a'4  d''             c''4      b'8  a'

  g'4. d''8            cs''  d'' ef'' d''
  g'4. d''8            ef''4     d''
  g'4. d''8            cs''  d'' c''  d''
  bf'8 d'' a' d''      g'    d'' f'   d''

  ef'8 bf' a' bf'      c''   bf' a'   bf'
  d'8  bf' a' bf'      c''   bf' a'   bf'
  df'8 bf' a' bf'      c''   bf' a'   bf'
  c'8  bf' a' g'       a'    fs' g'   gs'

  \bar "||" %% Bar  25

  a'8  e'' ds'' e''   f''  e'' ds'' e''
  a'' e'' ds'' e''    c''' e'' ds'' e''

  g'8  e'' ds'' e''   f''  e'' ds'' e''
  b'' e'' ds'' e''    a''  e'' ds'' e''

  f'8  e'' ds'' e''   f''  e'' ds'' e''
  a'' e'' ds'' e''    g''  e'' ds'' e''

  e'8  e'' ds'' e''   e'   d'' cs'' d''
  e'8  c'' b'   c''   e'   b'  as'  b'

  e'8  a' c'' e''     f'' e'' d'' c''
  b'8  d' g'  b'      d''2
  c''8 a' c'' e''     f'' e'' d'' c''
  b'8  e' gs' b'      d''2

  \bar "||" %% Bar  37

  << { f'1  } \\ { a'4   d''  c''  b' } >>
  << { g'1  } \\ { a'4   d''  c''  b' } >>
  << { e'1  } \\ { a'4   d''  c''  b' } >>
  << { f'1  } \\ { bf'4  d''  c''  bf' } >>
  << { fs'1 } \\ { b'4   d''  cs'' b' } >>
  << { g'1  } \\ { c''4  e''  d''  c'' } >>
  << { g'1  } \\ { b'4   g''  fs'' e'' } >>
  << { a'1  } \\ { ds''4 fs'' e''  ds'' } >>

  \bar "||" %% Bar  45

  <<
    { d''8 c'' c'' d'' d'' c'' c'' d''} \\
    { <f' a'>2 q }
  >>
  <<
    { d''8 c'' c'' d'' d'' c'' c'' d''} \\
    { <e' gs'>2 q }
  >>
  <<
    { d''8 c'' c'' d'' d'' c'' c'' d''} \\
    { <fs' as'>2 q }
  >>
  <<
    { d''8 c'' c'' d'' d'' c'' c'' d''} \\
    { <fs' a'>2 q }
  >>
  <<
    { d''8 c'' c'' d'' d'' c'' c'' d''} \\
    { <e' g'>2 q }
  >>
  <<
    { d''8 c'' c'' d'' d'' c'' c'' d''} \\
    { <d' fs'>2 q }
  >>
  <<
    { d''8 c'' c'' d'' d'' c'' c'' d''} \\
    { <d' f'>2 q }
  >>
  <<
    { d''8 c'' c'' d'' c'' b' b' c''} \\
    { <e' a'>2 <e' gs'> }
  >>

  \bar "||" %% Bar  53

  <<
    { b'2~ b'8 a' gs' a' } \\
    { <d' gs'>2 s2 }
  >>
  { <f' a' d''>4 <e' g' c''> <d' gs' b'> <c' f' a'> }
  <<
    { b'2~ b'8 f' e' f' } \\
    { <b e' gs'>2 s2 }
  >>
  { <b? ef' g'>4 <b d' f'> <gs c' e'> <gs b d'> }

  <gs b  e'>2~         e'8 e'' ds'' e''
  <a  d' f'>2~         f'8 e'' ds'' e''
  fs'8 e'' ds'' e''    g'  e'' gs'  e''

  \bar "||" %% Bar  60

  a'4. e''8            ds''  e'' f''  e''
  a'4. e''8            f''4      e''
  a'4. e''8            ds''  e'' ds'' e''
  a'4  d''             c''4      b'8  a'

  g'4. d''8            cs''  d'' ef'' d''
  g'4. d''8            ef''4     d''
  g'4. d''8            cs''  d'' c''  d''
  bf'8 d'' a' d''      g'    d'' f'   d''

  ef'4. bf'8           a'    bf' a'   bf'
  ef'4. bf'8           c''4      bf'
  d'4.  bf'8           a'    bf' a'   bf'
  c'8  bf' a' g'       a'    fs' g'   gs'

  \bar "||" %% Bar  72

  a'8 c'' b' c''   e'' c'' b'  c''
  g'8 c'' b' c''   e'' c'' b'  c''
  f'8 c'' b' c''   f'' c'' b'  c''
  e'8 c'' b' c''   e'  b'  as' b'

  e'8 fs' gs' a' b' c'' d'' e''
  gs'8 a' b' c'' d'' e'' fs'' gs''
  <c'' e'' a''>1
  <c' e' a'>1

  \bar "|."
}

the_lower = {
  \set Staff.midiInstrument = "harpsichord"
  \clef bass
  \key a \minor
  \time 4/4

  a8 c' e' r r2
  a8 g  f  r r2
  a8 c' e' r r2
  a8 g  f  r r2
  \break

  g8 bf d' r r2
  g8 f  ef r r2
  g8 bf d'4  cs' c'
  bf    a    g   f
  \break

  ef8 g bf r r2
  ef8 d c  r r2
  d8  g bf r r2
  c8  g a  r  r2
  \break

  \bar "||" %% Bar  13

  a,4 <a c' e'>8 q   <a c' ds'>4 <a c' f'>
  a,4 <a c' e'>8 q   <a d' f' >4 <a c' e'>
  a,4 <a c' e'>8 q   <a c' ds'>4 <a c' ds'>
  a,4 <a b  d'>8 q   <e a  c' >4 <e gs b>
  \break

  g,4 <g bf d'>8 q   <g bf cs'>4 <g bf ef'>
  g,4 <g bf d'>8 q   <g c' ef'>4 <g bf d'>
  g,4 <g bf d'>8 q   <g bf cs'>4 <g bf c'>
  <d bf>4 <d a>      <d g> <d f>
  \break

  <bf, ef>4 <ef g bf>8 q   <ef g c'>4 <ef g a>
  <bf, d >4 <d  g bf>8 q   <d  g c'>4 <d  g a>
  <bf, df>4 <df g bf>8 q   <df g c'>4 <df g a>
  <a, c>4   <ef fs a>8 q   <d fs a >4 <d  g a>
  \break

  \bar "||" %% Bar  25
  \pageBreak

  <c  e   a >2   <a  d'  f'>
  <a  c'  e'>2   <e  a   c'>
  <c  e   g >2   <f  a   c'>
  <e  g   b >2   <c  e   a >
  \break

  <a, c   f >2   <f  a   c'>
  <c  f   a >2   <b, e   g >
  <e, b,  e >2   <e, b,  d >
  <e, gs, c >2   <e, gs, b,>
  \break

  <a, c   e >4   <c   e  a>8 q   <c   e  a>4   <a, c   e >8 q
  <g, b,  d >4   <b,  d  g>8 q   <b,  d  g>4   <g, b,  d >8 q
  <f, a,  c >4   <a,  c  f>8 q   <a,  c  f>4   <f, a,  c >8 q
  <e, gs, b,>4   <gs, b, e>8 q   <gs, b, e>4   <e, gs, b,>8 q
  \break

  \bar "||" %% Bar  37

  <f,  c  f>1
  <g,  d  g>1
  <a,  e  a>1
  <bf, f  bf>1
  \break

  <b,  fs b>1
  <c   g  c'>1
  <b,  g  b>1
  <b,  fs b>1
  \break

  \bar "||" %% Bar  45

  <f, c>2 <c f> <e, b,> <b, e> <ds, as,> <as, ds> <d, a,> <a, d>
  \break

  <c, g,>2 <g, c> <b,, fs,> <fs, b,> <bf,, f,> <f, bf,> <a,, e,> <e, gs,>
  \break

  \bar "||" %% Bar  53
  \pageBreak

  e,8 gs, b, e    gs2
  f,8 d   e f     gs f e d
  e,8 gs, b, e    gs2
  ef,8 g, b,? d   e d b, f,
  \break

  e,8 gs, b, e   gs2
  f,8 a,  d  f   a2
  <a, c ef fs>2 <c ef g>4 <c ef gs>4

  \bar "||" %% Bar  60

  a,8 c e a    c'4 b
  a8 g f4      e8 d c b,
  a,8 c e a    c'4 b
  a8 g f e     d8 c b, a,
  \break

  g,8 bf, d g  bf4 a
  g8 f ef4     d8 c bf, a,
  g,8 bf, d g  bf4 a
  <g, g>4 <f, f> <ef, ef> <d, d>
  \break

  <ef, ef>4 <bf, g> <c g> <bf, g>
  <ef, ef>4 <bf, g> <c g> <bf, g>
  <d, d>4   <bf, g> <c g> <bf, g>
  <c, c>4   <bf, g> <c g> <bf, g>
  \break

  \bar "||" %% Bar  72

  <a, e a>4 <e  b>8 q   <e c'>4 <e   b>
  <g, e g>4 <e  b>8 q   <e c'>4 <e   b>
  <f, e f>4 <e  b>8 q   <e c'>4 <e   b>
  <e,   e>4 <b, g>      <e, e>4 <as, g>
  \break

  <e, b, e >1
  <e  b  e'>1
  a4 e c e
  <a, e  a >1

  \bar "|."
}

\score {
  <<
    \new PianoStaff <<
      \new Staff = "upp" \the_upper
      \new Staff = "low" \the_lower
    >>
  >>

  \layout { }
  \midi { }
}
