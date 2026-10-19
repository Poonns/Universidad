function [raiz, iter, errorAprox] = newton(x0, f, df, maxiteraciones, tol)
%NEWTON Calcula una raiz usando el metodo de Newton.

if nargin < 5
    tol = 1e-6;
end

if nargin < 4
    maxiteraciones = 100;
end

raiz = x0;
errorAprox = inf;

for iter = 1:maxiteraciones
    fx = f(raiz);
    dfx = df(raiz);

    if dfx == 0
        error('La derivada es cero. No se puede continuar.');
    end

    xNuevo = raiz - fx / dfx;
    errorAprox = abs(xNuevo - raiz);
    raiz = xNuevo;

    if errorAprox < tol
        return
    end
end

warning('Se alcanzo el maximo de iteraciones.');
end
