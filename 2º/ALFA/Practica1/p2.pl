inicial(a).
final(f).

%Reglas
af(S, [], S).

af(State, [X | Resto], End) :- 
    d(State, X, NextState), 
    af(NextState, Resto, End).

aceptar(Cadena) :- 
    af(Inicial, Cadena, Final),
    inicial(Inicial), 
    final(Final).

%Estado A
d(a, 1, b). d(a, 0, a). d(a, 2, a).

%Estado B
d(b, 2, c). d(b, 0, b). d(b, 1, b).

%Estado C
d(c, 0, d). d(c, 1, c). d(c, 2, c).

%Estado D
d(d, 0, e). d(d, 1, d). d(d, 2, d).

%Estado E
d(e, 1, f). d(e, 0, e). d(e, 2, e).

%Estado F
d(f, 0, f). d(f, 1, f). d(f, 2, f).
