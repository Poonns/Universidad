function [x, iter, errorJ] = jacobi(A, b, x0, max_iter, tol)
D = diag(diag(A)); 
N = D - A;         
invD = inv(D);

%matriz de iteración de Jacobi
QJ = invD * N;

%tamaño de la matriz de iteración
tam = norm(QJ, inf);

x = x0; 
errorJ = 0; 

for iter = 1:max_iter
    x_it = invD * b + QJ * x;

    errorJ = (tam / (1 - tam)) * norm(x_it - x, inf);  

    %comprobación de la tolerancia
    if errorJ < tol
        x = x_it;
        return;
    end

    x = x_it;
end

disp('Jacobi: Se alcanzó el número máximo de iteraciones sin que converja.');
end