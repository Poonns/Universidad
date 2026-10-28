function [x, iter, errorG] = gauss(A, b, x0, max_iter, tol)
M = tril(A); 
N = M - A;         
invM = inv(M);

%matriz de iteración de Gauss-Seidel
QG = invM * N;

%tamaño de la matriz de iteración 
tam = norm(QG, inf);

x = x0; 
errorG = 0; 

for iter = 1:max_iter
    x_it = invM * b + QG * x;

    errorG = (tam / (1 - tam)) * norm(x_it - x, inf);

    %comprobación de la tolerancia
    if errorG < tol
        x = x_it;
        return;
    end

    x = x_it;
end

disp('Gauss-Seidel: Se alcanzó el número máximo de iteraciones sin que converja.');
end