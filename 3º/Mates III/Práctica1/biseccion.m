function [ap,nit,cota] = biseccion(a,b,f,maxit,tol)
%BISECCION Aproxima una raíz de una función por el método de bisección.
%   Entradas:
%       A, B   Extremos del intervalo inicial.
%       F      Función anónima o handle.
%       MAXIT  Número máximo de iteraciones.
%       TOL    Tolerancia para detener el algoritmo.
%
%   Salidas:
%       AP     Aproximación de la raíz.
%       NIT    Número de iteraciones realizadas.
%       COTA   Cota del error estimada.
%
%   Nota:
%       El método requiere que F(A) y F(B) tengan signos opuestos.
    
    %Inicializacion variables
    fa = f(a);
    fb = f(b);
    nit = 0;
    cota = (b - a)/2;

    if fa * fb > 0
        error('La función no cumple el teorema del valor intermedio en el ')
    end

    % Algoritmo de la bisección
    while abs(cota) > tol && nit < maxit
        nit = nit + 1;
        c = (a + b) / 2;
        fc = f(c);
        
        %Comporbación del cambio de signo
        if fa * fc < 0
            b = c;
            fb = fc;
        else
            a = c;
            fa = fc;
        end

        cota = cota /2;
    end

    ap = (a + b)/2;
    if nit == maxit
       warning('Advertencia: Se alcanzço el numero maximo de iteraciones')
    end
end