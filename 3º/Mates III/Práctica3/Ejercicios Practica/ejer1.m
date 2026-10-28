%definimos la matriz y el vector
A = [12, -1,  0,  0; 
    -2, 11, -1,  0; 
    0, -3, 10, -1; 
    0,  0, -1,  9];
b = [5; 15; 15; 10];

%a)
diagonal = abs(diag(A));
suma_resto_fila = sum(abs(A), 2) - diagonal; 

if all(diagonal > suma_resto_fila)
    disp('La matriz es estrictamente diagonalmente dominante');
else
    disp('La matriz no es estrictamente diagonalmente dominante');
end

%definimos los parametros
x0 = zeros(4, 1); 
tol = 1e-5;       
max_iter = 100;   

%b), c)
[x_J, iter_J, err_J] = jacobi(A, b, x0, max_iter, tol);
disp(['Jacobi ha terminado en ', num2str(iter_J), ' iteraciones. Con un error de: ', num2str(err_J)]);
disp('Solución aproximada Jacobi:');
disp(x_J);

[x_G, iter_G, err_G] = gauss(A, b, x0, max_iter, tol);
disp(['Gauss-Seidel ha terminado en ', num2str(iter_G), ' iteraciones. Con un error de: ', num2str(err_G)]);
disp('Solución aproximada Gauss:');
disp(x_G);

% d)
x_exacta = A \ b; %resolucion directa
disp('Solución exacta (A \ b):');
disp(x_exacta);