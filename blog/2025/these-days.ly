\version "2.24.1"
\include "english.ly"
\pointAndClickOff

\header {
  title = "These Days (One Night Lonely)"
  composer = \markup \right-column {
    "by Powderfinger"
    "arranged by Nigel Tao"
    "for Piano (and optional Vocals)"
  }
  tagline = #f
}

the_chords = \chords {
  \tempo 4 = 90

  { a1 fs:m e d } \break

  { a1 fs:m e d } \break
  { a1 fs:m e d2 d:m } \break

  { a2 e fs:m d e1 } \break
  { a2 e fs:m d e1 } \break

  { a1 fs:m e d } \break
  { a1 fs:m e d } \break
  { a1 fs:m e d } \break
  { a1 fs:m e d2 d:m } \break

  { a2 e fs:m d e1 } \break
  { a2 e fs:m d e1 } \break
  { a2 e fs:m d e1 } \break
  { a2 e fs:m d e1 e:7 } \break

  { a2 e:7 a e a2 e:7 a e } \break

  { d1 a2 e d1 } \break
  { a2 e d1 d:m } \break

  { a fs:m e d } \break
  { a fs:m e d d:m } \break

  { a2 e fs:m d e1 } \break
  { a2 e fs:m d e1 } \break \pageBreak
  { a2 e fs:m d e1 } \break
  { a2 e fs:m d e1 } \break

  { a2 e fs:m d e1 } \break
  { a2 e fs:m d e1 } \break \pageBreak
  { a2 e fs:m d e1 } \break
  { a2 e fs:m d e1 } \break

  { a2 e fs:m d e1 } \break
  { a2 e fs:m d e2 d4 e a1 }
}

the_lyrics = \lyricmode {

  \set fontSize = #(magnification->font-size 0.75)

  It's co -- ming round a -- gain
  The slow -- ly cree -- ping hand
  Of time and its com -- mand
  Soon e -- nough it comes
  Set -- tles in its place
  Its shad -- ow in my face
  Puts pres -- sure in my day

  This life well it's slip -- ping right through my hands
  These days turned out no -- thing like I had planned

  It's co -- ming round a -- gain
  The slow -- ly cree -- ping hand
  Of time and its com -- mand
  It set -- tles in its place
  Its shad -- ow in my face
  Puts pres -- sure in my day

  Soon e -- nough it comes
  And here it is a -- gain
  Slow -- ly cree -- ping hand
  Time and its com -- mand

  Soon e -- nough it comes
  Set -- tles in its place
  Pres -- sure in my day
  Un -- dig -- ni -- fied and lame

  This life well it's slip -- ping right through my hands
  These days turned out no -- thing like I had planned

  Con -- trol well it's slip -- ping right through my hands
  These days turned out no -- thing like I had planned

  Ooh Ooh

  Soon e -- nough it comes
  Soon e -- nough it comes
  To tie us down

  Oh
  It's co -- ming round a -- gain
  Slow
  Slow -- ly cree -- ping hand
  Hand

  Na na na na na na na na na na na na na na
  Na na na na na na na na na na na na na na
  Na na na na na na na na na na na na na na
  Na na na na na na na na na na na na na na

  This life well it's slip -- ping right through my hands
  These days turned out no -- thing like I had planned
}

the_vocals = {
  \clef treble
  \key a \major
  \time 4/4

  { r1 r r r }

  | %% Bar  5.

  { r4.          a'8 d''  cs'' b' a' }
  { cs''4.       a'8 d''  cs'' b' a' }
  { b'4.         e'8 cs'' b'   a' b' }
  { a'2          fs''8 e''  b' a' }

  { cs''2        d''8  cs'' b' a' }
  { cs''4.       a'8 d''  cs'' b' a' }
  { b'4.         e'8 cs'' b'   a' b' }
  { a'2          r2 }

  | %% Bar 13.

  { r4 cs''      b'  b'8 cs'' }
  { d'' cs'' b' a'~ a'4 cs''}
  { b'1 }
  { r4 cs''      b'  b'8 cs'' }
  { d'' cs'' b' a'~ a'4 fs'}
  { e'1 }

  | %% Bar 19.

  { r4.    cs''8    d''  cs'' b'   a' }
  { cs''4. a'8      d''  e''  cs'' a' }
  { b'4.   e'8      cs'' b'   a'   b' }
  { a'4.  (fs'16 e' d'2) }

  | %% Bar 23.

  { r4.      a'8 d''   cs'' b'   a' }
  { cs''4.   a'8 d''   cs'' b'   a' }
  { b'4.     e'8 cs''  b'   cs'' b' }
  { a'2          fs''8 e''  b'   a' }

  | %% Bar 27.

  { cs''4.   a'8 d''   cs'' b'   a'  }
  { cs''2        d''8  e''  cs'' a'  }
  { b'2          cs''8 b'   a'   b'  }
  { a'2          fs''8 e''  fs'' e'' }

  | %% Bar 31.

  { cs''2        d''8  cs'' b' a' }
  { cs''2        d''8  cs'' b' a' }
  { b'4 e'       cs''8 b'   a' b' }
  { a'2          r2 }

  | %% Bar 35.

  { r4 cs''      b'  b'8 cs'' }
  { d'' cs'' b' a'~ a'4 cs''}
  { b'1 }
  { r4 cs''      b'  b'8 cs'' }
  { d'' cs'' b' a'~ a'4 cs''}
  { b'2 (e') }

  | %% Bar 41.

  { r4 cs''      b'  b'8 cs'' }
  { d'' cs'' b' a'~ a'4 cs''}
  { b'1 }
  { r4 cs''      b'  b'8 cs'' }
  { d'' cs'' b' a'~ a'4 cs''}
  { b'1 r1 }

  | %% Bar 48.

  { r1 a'2   (gs') }
  { r1 cs''2 (b')  }

  | %% Bar 52.

  { r2           fs''8 e''  b' a' }
  { cs''2        r2 }
  { r2           fs''8 e''  b' a' }
  { cs''4. cs''8 b'4 gs' }
  { a'1 r1 }

  | %% Bar 58.

  {           r2 cs'' (a'4.)   e'8 fs'  a' cs'' a' }
  { b'2 r2 r1 r2 cs'' (a')         fs'8 a' cs'' b' }
  { b'2 r2 }
  { r2. cs''8 (b' a'1) }

  | %% Bar 67.

  r1 r r
  r1 r r
  r1 r r
  r1 r r2. a'8 b'

  | %% Bar 79.

  { cs''4 d''8 cs'' b' cs'' b' a'~ a'4 fs'8 a' cs'' b'4 b'8~ b'2. a'8 b' }
  { cs''4 d''8 cs'' b' cs'' b' a'~ a'4 fs'8 a' cs'' b'4 b'8~ b'2. a'8 b' }
  { cs''4 d''8 cs'' b' cs'' b' a'~ a'4 fs'8 a' cs'' b'4 b'8~ b'2. a'8 b' }
  { cs''4 d''8 cs'' b' cs'' b' a'~ a'4 fs'8 a' cs'' b'4 b'8~ b'1 \fermata }

  | %% Bar 91.

  { r4 cs''      b'  b'8 cs'' }
  { d'' cs'' b' a'~ a'4 cs''}
  { b'1 }

  | %% Bar 94.

  { r4 cs''      b'  b'8 cs'' }
  { d'' cs'' b' a'~ a'4 cs''}
  { b'4 (e'2.) }
  { r1 \bar "|." }
}

the_upper = {
  \clef bass
  \key a \major
  \time 4/4

  r4 <e a cs'>4. e8 <a   b>   cs'
  a2          r8 a  <cs' fs'> a
  <gs e'>2    r8 e  <a   b>   b,
  <a fs'>2.         <b   e'>4

  | %% Bar  5.

  \clef treble

  { <a cs'>4.  a'8   d''   cs'' b' a' }
  { cs''4.     a'8   d''   cs'' b' a' }
  { b'4.       e'8   cs''  b'   a' b' }
  { a'2              fs''8 e''  b' a' }

  | %% Bar  9.

  { cs''2           d''8 cs'' b' a' }
  { cs''4. a'8      d''  cs'' b' a' }
  { b'4.   e'8      cs'' b'   a' b' }
  { a'2             r2 }

  | %% Bar 13.

  r4 <<
    { cs''4      b'  b'8 cs'' } \\
    { <e' a'>4 <e' gs'> q }
  >> <<
    { d''8 cs'' b' a'~ a'4 <fs' cs''>} \\
    { fs'2 d' }
  >>
  { <e' gs' b'>1 }

  | %% Bar 16.

  r4 <<
    { cs''4      b'  b'8 cs'' } \\
    { <e' a'>4 <e' gs'> q }
  >> <<
    { d''8 cs'' b' a'~ a'4 <d' fs'>} \\
    { fs'2 d' }
  >>
  { <gs b e'>2 <a d' fs'>4 <b e' gs'> }

  | %% Bar 19.

  <cs' e' a'>4   << { s8 cs'' d'' cs'' b' a' }  \\ { <e'  a'>4 q q } >>
  <fs' a' cs''>4 << { s8 a'8 d'' e'' cs'' a' }  \\ { <fs' a'>4 q q } >>
  <e' gs' b'>4   << { s8 e' cs'' b' a' b' }     \\ { <e' gs'>4 q q } >>
  <d' fs' a'>4. fs'16 e' <a d'>4 \grace <a e'>16 <a fs'>4

  | %% Bar 23.

  <a cs'>4       << { s8 a' d'' cs'' b' a' }  \\ { <cs' e'>4 <e' a'> q } >>
  <fs' a' cs''>4 << { s8 a' d'' cs'' b' a' }  \\ { <fs' a'>4 q q } >>
  <e' gs' b'>4   << { s8 e' cs'' b' cs'' b' } \\ { <e' gs'>4 q q } >>
  <d' fs' a'>2 <fs' fs''>8 <e' e''> <b b'> <a a'>

  | %% Bar 27.

  <cs' cs''>4 <<
    { s8 a' d'' cs'' b' a' } \\
    { <cs' e'>4 <e' a'> <e' gs'> } >>
  <fs' a' cs''>4 <fs' a'> <<
    { d''8 e'' cs'' a' } \\
    { <fs' a'>4 <e' gs'> } >>
  <e' gs' b'>4 <e' gs'> <<
    { cs''8 b' a' b' } \\
    { <e' gs'>4 <d' fs'> } >>
  <d' fs' a'>2 <fs' fs''>8 <e' e''> <fs' fs''> <e' e''>

  | %% Bar 31.

  <cs' cs''>4 <cs' e' a'> <d' d''>8 <cs' cs''> <b b'> <a a'>
  <cs' cs''>4 <cs' fs' a'> <d' d''>8 <cs' cs''> <b b'> <a a'>
  <b b'>4 <gs b e'> <cs' cs''>8 <b b'> <a a'> <b b'>
  <a a'>1

  | %% Bar 35.

  r4 <<
    { cs''4      b'  b'8 cs'' } \\
    { <e' a'>4 <e' gs'> q }
  >> <<
    { d''8 cs'' b' a'~ a'4 <fs' cs''>} \\
    { fs'2 d' }
  >>
  { <e' gs' b'>1 }

  | %% Bar 38.

  r4 <<
    { cs''4      b'  b'8 cs'' } \\
    { <e' a'>4 <e' gs'> q }
  >> <<
    { d''8 cs'' b' a'~ a'4 <fs' cs''>} \\
    { fs'2 d' }
  >>
  { <e' gs' b'>2 e' }

  | %% Bar 41.

  r4 <<
    { cs''4      b'  b'8 cs'' } \\
    { <e' a'>4 <e' gs'> q }
  >> <<
    { d''8 cs'' b' a'~ a'4 <fs' cs''>} \\
    { fs'2 d' }
  >>
  { <e' gs' b'>1 }

  | %% Bar 44.

  r4 <<
    { cs''4      b'  b'8 cs'' } \\
    { <e' a'>4 <e' gs'> q }
  >> <<
    { d''8 cs'' b' a'~ a'4 <fs' cs''>} \\
    { fs'2 d' }
  >>
  { <e' gs' b'>2 <gs b e'>8 q4 q8 }
  { <gs b d'>4 q8 q4. q4 }

  | %% Bar 48.

  << { e'4 d'8 cs' d' cs'4 cs'8 } \\ { cs'2 b } >>
  << { a'2 gs' }                  \\ { cs'4 cs' cs'8 b b4 } >>
  << { e'4 d'8 cs' d' cs'4 cs'8 } \\ { cs'2 b } >>
  << { cs''2 b' }                 \\ { cs'2 b2 } >>

  | %% Bar 52.

  <<
    { e'8 fs' a'4 fs''8 e'' b' a' } \\
    { d'2 a' }
  >> <<
    { cs''2 s } \\
    { b'8 a'4 a'8~ a' gs'4 e'8 }
  >> <<
    { e'8 fs' a'4 fs''8 e'' b' a' } \\
    { d'2 a' }
  >> <<
    { cs''4. cs''8 } \\
    { b'8 a'4 a'8 }
  >> <gs' b'>4 <e' gs'>

  << { e'4 d'8 cs' d' cs'4 cs'8 } \\ { fs1 } >>
  { <f d'>8 <e cs'> <f d'> <g e'> <f d'> <e cs'>~ q4 }

  | %% Bar 58.

  r2 <e' a' cs''>4. b'8
  <cs' fs' a'>4. e'8 fs' a' cs''16 b' a'8
  <e' gs' b'>2. fs'4
  <a d' fs'>1

  | %% Bar 62.

  r2 <e' a' cs''>4. b'8
  <cs' fs' a'>2 fs'8 a' cs'' b'
  <e' gs' b'>1
  r2. cs''8 b'
  <d' f' a'>1

  | %% Bar 67.

  <cs' e' a'>2 <b e' gs'>
  << { a'8 gs'4 fs'8~ fs'4 } \\ { <cs' fs'>2. } >> <a d' fs'>4 <a b fs'>8 <gs b e'>4. q2

  | %% Bar 70.

  <cs' e' a'>2 <b e' gs'>
  << { a'8 gs'4 fs'8~ fs'4 } \\ { <cs' fs'>2. } >> <a d' fs'>4 <a b fs'>8 <gs b e'>4. q2

  | %% Bar 73.

  <e' a' cs''>2 <e' gs' b'>
  << { a'8 gs'4 fs'8~ fs'4 } \\ { <cs' fs'>2. } >> <d' fs' cs''>4 <e' gs' cs''>8 <e' gs' b'>4. { cs''8 e'' cs'' b' }

  | %% Bar 76.

  <e' a' cs''>2 <e' gs' b'>
  << { cs''8 b'4 a'8~ a'4 } \\ { <fs' a'>2. } >> <d' fs' cs''>4 <e' gs' cs''>8 <e' gs' b'>4. <gs' b' e''>4 a'8 b'

  | %% Bar 79.

  { cs''4 d''8 cs'' b' cs'' b' a'~ a'4 fs'8 a' cs'' b'4 b'8~ b'2. a'8 b' }

  | %% Bar 82.

  { cs''4 d''8 cs'' b' cs'' b' a'~ a'4 fs'8 a' cs'' b'4 b'8~ b'2. a'8 b' }

  | %% Bar 85.

  << { cs''4 d''8 cs'' b' cs'' b' a'~ a' gs' fs' a' cs'' b'4 b'8 } \\ { <e' a'>2 <e' gs'> cs' d'4 fs' } >>
  { <e' gs' b'>4 e''16 fs'' e'' fs'' e''4 a'8 b' }

  | %% Bar 88.

  << { cs''4 d''8 cs'' b' cs'' b' a'~ a' gs' fs' a' cs'' b'4 b'8 } \\ { <e' a'>2 <e' gs'> cs' d'4 fs' } >>
  { <e' gs' b'>1 \fermata }

  | %% Bar 91.

  r4 <<
    { cs''4      b'  b'8 cs'' } \\
    { <e' a'>4 <e' gs'> q }
  >> <<
    { d''8 cs'' b' a'~ a'4 <fs' cs''>} \\
    { fs'2 d' }
  >>
  { <e' gs' b'>1 }

  | %% Bar 94.

  r4 <<
    { cs''4      b'  b'8 cs'' } \\
    { <e' a'>4 <e' gs'> q }
  >> <<
    { d''8 cs'' b' a'~ a'4 <fs' cs''>} \\
    { fs'2 d' }
  >>
  { <e' gs' b'>4 e' <a d' fs'> <b e' gs'> }
  < cs' e' a'>1
}

the_lower = {
  \clef bass
  \key a \major
  \time 4/4

  a,1
<<
  { r4 <cs fs>2. } \\
  { fs,1 }
>> <<
  { r4 <b, e>2. } \\
  { e,1 }
>>
  { d,8 a, d a d2 }

  | %% Bar  5.

<<
  { r4 <cs e a> q q } \\
  { a,1 }
>> <<
  { r4 <fs a cs'> q q } \\
  { fs,1 }
>> <<
  { r4 <e gs b> q q } \\
  { e,1 }
>> <<
  { r4 <d fs a> q q } \\
  { d,1 }
>>

  | %% Bar  9.

<<
  { r4 <a cs' e'> q q } \\
  { a,1 }
>> <<
  { r4 <a cs' fs'> q q } \\
  { fs,1 }
>> <<
  { r4 <gs b e'> q q } \\
  { e,1 }
>>
  { d,8 a, fs a, f, d a d }

  | %% Bar 13.

  { a,8 e a e       e, b, e b, }
  { fs,8 cs fs cs   d, a, d a, }
  { e,8 b, e b,     fs, fs gs, gs }

  | %% Bar 16.

  { a,8 e a e       e, b, e b, }
  { fs,8 cs fs cs   d, a, d a, }
  { e,8 b, e b,     fs, fs gs, gs }

  | %% Bar 19.

  { a,8  e a e     a, e cs' e }
  { fs, cs fs cs   fs, cs a cs }
  { e, b, e b,     e, b, gs b, }
  { d, a, d a,     d, a, fs a, }

  | %% Bar 23.

  { a,8  e a e     a, e cs' e }
  { fs, cs fs cs   fs, cs a cs }
  { e, b, e b,     e, b, gs b, }
  { d, a, d a,     d, a, fs a, }

  | %% Bar 27.

  { <a,, a,>4 a, <a,, a,> <gs,, gs,> }
  { <fs,, fs,>4 fs, <fs,, fs,> <fs,, fs,> }
  { <e,, e,>4 e, <e,, e,> <e,, e,> }
  { <d,, d,>4 d, <d,, d,> <d,, d,> }

  | %% Bar 31.

  { <a,, a,>4 <a, a> <a,, a,> <gs,, gs,> }
  { <fs,, fs,>4 <fs, fs> <fs,, fs,> <fs,, fs,> }
  { <e,, e,>4 <e, e> <e,, e,> <e,, e,> }
  { <d,, d,>4 <d fs a> <d, d> <d f a> }

  | %% Bar 35.

  { a,8 e a e       e, b, e b, }
  { fs,8 cs fs cs   d, a, d a, }
  { e,8 b, e b,     e, b, fs, gs, }

  | %% Bar 38.

  { a,8 e a e       e, b, e b, }
  { fs,8 cs fs cs   d, a, d a, }
  { e,8 b, e b,     e, b, fs, gs, }

  | %% Bar 41.

  { a,8 a e a       e, e b, e }
  { fs,8 fs cs fs   d, d a, d }
  { e,8 e b, e      gs, b, fs, gs, }

  | %% Bar 44.

  { a,8 e a e       e, b, e b, }
  { fs,8 cs fs cs   d, a, d a, }
  { e,8 b, e b,     e, b, e b, }
  { e,8 b, d b,     e, b, d b, }

  | %% Bar 48.

  { a,,8 e, a, e,    e, b, d b, }
  { a,8 e a e        e, b, e b, }
  { a,,8 e, a, e,    e, b, d b, }
  { a,8 e a e        e, b, e b, }

  | %% Bar 52.

  { << {d4. d8} \\ { d,2 } >> <d, d>2 }
  { a,8 e a4 e,8 b, e4 }
  { << {d4. d8} \\ { d,2 } >> <d, d>2 }
  { a,8 e a4 e,8 b, e4 }

  { <d, d>1 d, }

  | %% Bar 58.

  { << { r8 e a r r2 } \\ a,1 >> fs,1 e, d, }

  | %% Bar 62.

  { a,1 fs, e, d, d4 d d d }

  | %% Bar 67.

  { a,8 e a e      e, b, e b, }
  { fs,8 cs fs cs  d, a, d a, }
  { e,8 b, e b,    e, b, e b, }

  | %% Bar 70.

  { a,8 e a e      e, b, e b, }
  { fs,8 cs fs cs  d, a, d a, }
  { e,8 b, e b,    e, b, e b, }

  | %% Bar 73.

  { a,8 e a e      e, b, e b, }
  { fs,8 cs fs cs  d, a, d a, }
  { e,8 b, e b,    e, b, e b, }

  | %% Bar 76.

  { a,8 e a e      e, b, e b, }
  { fs,8 cs fs cs  d, a, d a, }
  { e,8 b, e b,    e, b, e b, }

  | %% Bar 79.

  { a,8 e a4       e,8 b, e4 }
  { fs,8 cs fs4    d,8 a, d4 }
  { e,8 b, e4      e,8 b, e4 }

  | %% Bar 82.

  { a,8 e a4       e,8 b, e4 }
  { fs,8 cs fs4    d,8 a, d4 }
  { e,8 b, e4      e,8 b, e4 }

  | %% Bar 85.

  { a,8 e a e      e, b, e b, }
  { fs,8 cs fs cs  d, a, d a, }
  { e,8 b, e b,    e, b, e b, }

  | %% Bar 88.

  { a,8 e a e      e, b, e b, }
  { fs,8 cs fs cs  d, a, d a, }
  { e,1 \fermata }

  | %% Bar 91.

  { a,2 e, fs, d, e, e }

  | %% Bar 94.

  { a,2 e, fs, d, e, fs,4 gs, a,1 }
}

\score {
  <<
    \the_chords
    \new Staff \with { instrumentName = "Vocals" }
	  { \new Voice = "voc" { \the_vocals } }
    \new Lyrics \lyricsto voc \the_lyrics
    \new PianoStaff \with { instrumentName = "Piano" } <<
      \new Staff \the_upper
      \new Staff \the_lower
    >>
  >>

  \layout { }
  %% \midi { }
}
