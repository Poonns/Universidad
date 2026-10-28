A3 = [ 1,  3, -4,  4; 
    0,  3, -4, -4; 
    4, -1,  0,  1; 
    -2,  1,  3,  0];
b3 = [-3; -1; -2;  2];

%a)
x_exacta3 = A3 \ b3;
disp('Solución exacta (A \ b):');
disp(x_exacta3);

%b)

%aplicamos jacobi
[x_J3, iter_J3, err_J3] = jacobi(A3, b3, x0, max_iter, tol);
error_real_J3 = norm(x_exacta3 - x_J3, inf); 

%aplicamos gauss
[x_G3, iter_G3, err_G3] = gauss(A3, b3, x0, max_iter, tol);
error_real_G3 = norm(x_exacta3 - x_G3, inf); 

fprintf('Método\t\t| Iteraciones\t| Cota de Error\t| Error Real\n');
fprintf('------------------------------------------------------------------\n');
fprintf('Jacobi\t\t|\t%d\t| %e\t| %e\n', iter_J3, err_J3, error_real_J3);
fprintf('Gauss-Seidel\t|\t%d\t| %e\t| %e\n\n', iter_G3, err_G3, error_real_G3);

% c) Significado de los valores y justificación
disp('--- Entrega 3: c) Justificación de los resultados ---');
disp('Los resultados obtenidos (NaN o Inf) indican que los métodos han fallado.');
disp('El motivo es que la matriz A tiene ceros en su diagonal principal (a33 = 0 y a44 = 0).');
disp('La formulación de los métodos de Jacobi y Gauss-Seidel requiere dividir por');
disp('los elementos de la diagonal principal para calcular cada nueva iteración. Al ser');
disp('cero, la matriz M utilizada en cada método no es invertible y el cálculo colapsa.');