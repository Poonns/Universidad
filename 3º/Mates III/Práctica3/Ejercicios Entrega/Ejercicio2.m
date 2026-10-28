A2 = [ 6,  3, -4,  4; 
    0,  8, -4, -4; 
    4, -1,  7,  1; 
    -2,  1,  3,  9];
b2 = [-3;  1; -2; -2];

%a)
x_exacta2 = A2 \ b2;
disp('Solución exacta (A \ b):');
disp(x_exacta2);

%b)

%aplicamos jacobi
[x_J2, iter_J2, err_J2] = jacobi(A2, b2, x0, max_iter, tol);
error_real_J2 = norm(x_exacta2 - x_J2, inf); 

%aplicamos gauss
[x_G2, iter_G2, err_G2] = gauss(A2, b2, x0, max_iter, tol);
error_real_G2 = norm(x_exacta2 - x_G2, inf); 

fprintf('Método\t\t| Iteraciones\t| Cota de Error\t| Error Real\n');
fprintf('------------------------------------------------------------------\n');
fprintf('Jacobi\t\t|\t%d\t| %e\t| %e\n', iter_J2, err_J2, error_real_J2);
fprintf('Gauss-Seidel\t|\t%d\t| %e\t| %e\n\n', iter_G2, err_G2, error_real_G2);

%c)
diagonal2 = abs(diag(A2));
suma_resto_fila2 = sum(abs(A2), 2) - diagonal2; 

if all(diagonal2 > suma_resto_fila2)
    disp('La matriz es estrictamente diagonalmente dominante.');
else
    disp('La matriz no es estrictamente diagonalmente dominante.');
end
