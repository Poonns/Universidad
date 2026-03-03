
inicial(a).
final(b).

%Reglas
af(S, [], S).

af(State, [Simbolo | Resto], End) :- 
    d(State, Simbolo, NextState), 
    af(NextState, Resto, End).

% Regla Épsilon
af(State, Cadena, End) :- 
    vacia(State, NextState), 
    af(NextState, Cadena, End).

aceptar(Cadena) :- 
    af(Inicial, Cadena, Final),
    inicial(Inicial), 
    final(Final).

% Transición épsilon (Del inicial 'a' al final 'b')
vacia(a, b).

% Transiciones a,b,c

d(b, a, a).     
d(a, a, c).     
d(c, b, c).     
d(c, a, b).     
d(c, b, b).     