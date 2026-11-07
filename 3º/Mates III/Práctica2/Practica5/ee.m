function [x, y, N] = ee(f, x0, y0, h, xf)
%generacion de la malla computacional
x = x0 : h : xf;
N = length(x) - 1; %numero de intervalos

%inicializacion del vector y
y = zeros(1, N + 1);
y(1) = y0;

%bucle iterativo metodo euler
for i = 1:N
    y(i+1) = y(i) + h * f(x(i), y(i));
end
end
