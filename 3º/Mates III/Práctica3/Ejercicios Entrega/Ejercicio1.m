A1 = [19,  3, -4,  4; 
    0, 17, -4, -4; 
    4, -1, 10,  1; 
    -2,  1,  3,  7];
b1 = [3; 1; 2; -2];

%a)
x_exacta1 = A1 \ b1;
disp('Solución exacta (A \ b):');
disp(x_exacta1);

x0 = zeros(4, 1);  
tol = 1e-4;      
max_iter = 20;     

%b)

%aplicamos jacobi
[x_J1, iter_J1, err_J1] = jacobi(A1, b1, x0, max_iter, tol);
error_real_J1 = norm(x_exacta1 - x_J1, inf); 

%aplicamos gauss
[x_G1, iter_G1, err_G1] = gauss(A1, b1, x0, max_iter, tol);
error_real_G1 = norm(x_exacta1 - x_G1, inf); 


fprintf('Método\t\t| Iteraciones\t| Cota de Error\t| Error Real\n');
fprintf('------------------------------------------------------------------\n');
fprintf('Jacobi\t\t|\t%d\t| %e\t| %e\n', iter_J1, err_J1, error_real_J1);
fprintf('Gauss-Seidel\t|\t%d\t| %e\t| %e\n\n', iter_G1, err_G1, error_real_G1);
