inicial(a).
final(g). 
final(i). 

af(S, [], S).

af(State, [X | Resto], End) :- 
    d(State, X, NextState), 
    af(NextState, Resto, End).

aceptar(Cadena) :- 
    af(Inicial, Cadena, Final),
    inicial(Inicial), 
    final(Final).

%Prefijo
d(a, 0, b).
d(b, 1, c).
d(c, 1, d).

% Indeterminismo de D
d(d, 2, e).
d(d, 2, f). 

% Camino superior
d(e, 2, i).
d(i, 2, j).
d(j, 2, i).

%Camino inferior
% Estado F
d(f, 1, g).
d(f, 0, f).
d(f, 2, f).

% Estado G (FINAL)
d(g, 1, h).
d(g, 0, g).
d(g, 2, g).

% Estado H
d(h, 1, f).
d(h, 0, h).
d(h, 2, h).
