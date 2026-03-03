inicial(a).
final(f).
final(f1).
final(f2).

%Reglas
af(S, [], S).

af(State, [X | Resto], End) :- 
    d(State, X, NextState), 
    af(NextState, Resto, End).

aceptar(Cadena) :- 
    af(Inicial, Cadena, Final),
    inicial(Inicial), 
    final(Final).


%Primer escalón
d(a, 0, a1).  d(a, 1, b). d(a, 2, a1).
d(b, 0, a1).  d(b, 1, a1). d(b, 2, c).
d(c, 0, d).  d(c, 1, a1). d(c, 2, a1).
d(d, 0, e).  d(d, 1, a1). d(d, 2, a1).
d(e, 0, a1).  d(e, 1, f). d(e, 2, a1).

%Segundo escalón
d(a1, 0, a2). d(a1, 1, b1). d(a1, 2, a2).
d(b1, 0, a2).  d(b1, 1, a2). d(b1, 2, c1).
d(c1, 0, d1).  d(c1, 1, a2). d(c1, 2, a2).
d(d1, 0, e1).  d(d1, 1, a2). d(d1, 2, a2).
d(e1, 0, a2).  d(e1, 1, f1). d(e1, 2, a2).

%Tercer escalón
d(a2, 0, err). d(a2, 1, b2). d(a2, 2, err).
d(b2, 0, err).  d(b2, 1, err). d(b2, 2, err).
d(c2, 0, d2).  d(c2, 1, err). d(c2, 2, err).
d(d2, 0, e2).  d(d2, 1, err). d(d2, 2, err).
d(e2, 0, err).  d(e2, 1, f2). d(e2, 2, err).

%Error
d(err, 0, err). d(err, 1, err). d(err, 2, err).

%Final
d(f, 0, f).   d(f, 1, f).   d(f, 2, f).
d(f1, 0, f1). d(f1, 1, f1). d(f1, 2, f1).
d(f2, 0, f2). d(f2, 1, f2). d(f2, 2, f2).