inicial(a).
final(g).

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
d(a, 0, a). d(a, 1, c). d(a, 2, b).

%Estado B
d(b, 0, b). d(b, 1, b). d(b, 2, a).

%Estado C
d(c, 0, a). d(c, 1, a). d(c, 2, d).

%Estado D
d(d, 0, e). d(d, 1, b). d(d, 2, a).

%Estado E
d(e, 0, f). d(e, 1, b). d(e, 2, a).

%Estado F
d(f, 0, b). d(f, 1, g). d(f, 2, a).

%Estado G
d(g, 0, g). d(g, 1, g). d(g, 2, g).