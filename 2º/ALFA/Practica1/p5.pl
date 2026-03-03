inicial(a).
final(k).

af(S, [], S).

af(State, [X | Resto], End) :- 
    d(State, X, NextState), 
    af(NextState, Resto, End).

% Regla Épsilon 
af(State, Cadena, End) :- 
    vacia(State, NextState), 
    af(NextState, Cadena, End).

aceptar(Cadena) :- 
    af(Inicial, Cadena, Final),
    inicial(Inicial), final(Final).

% Camino de arriba 
d(a, 1, b). 
d(b, 2, c). 
d(c, 0, d). 
d(d, 0, e). 
d(e, 1, k). 

% Camino de abajo 
d(g, 2, h). 
d(h, 2, i). 
d(i, 2, j). 
d(j, 2, k).

% Transiciones épsilon 
vacia(a, g).
vacia(b, g).
vacia(c, g).
vacia(d, g).
vacia(e, g).

%Estado K
d(k, 0, k).
d(k, 1, k).
d(k, 2, k).
