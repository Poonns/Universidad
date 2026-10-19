f3 = @(x)  exp(-x) - x;
fd3 = @(x) -exp(-x) - 1;


val_a = f3(0);
val_b = f3(1);

fprintf('Evaluación en los extremos del intervalo [0, 1]:\n');
fprintf('f3(0) = %.4f \n', val_a);
fprintf('f3(1) = %.4f \n', val_b);
disp('Existe un cambio de signo, por lo que hay al menos una raíz.'); 

x = linspace(0, 1, 100); 
y = f3(x);
x0 = 0.5;

plot(x, y) 
yline(0, '--k') 
grid on 
xlabel('Eje X') 
ylabel('Eje Y') 
title('Raíz de f_3(x) = exp(-x) - x; en [0,1]') 
legend('f_3(x)', 'y = 0') 

tol = 1e-5;
maxiteraciones = 100;
%Calculo Via Biseccion%
[raiz_bis, iter_bis, error_bis] = biseccion(0, 1, f3, maxiteraciones, tol);
fprintf('Raíz aproximada: %.6f\n', raiz_bis);
fprintf('Iteraciones: %d\n', iter_bis); 
fprintf('Cota de error: %.6e\n', error_bis); 

%Calculo Via Newton%
[raiz_new, iter_new, error_new] = newton(x0, f3, fd3, maxiteraciones, tol); 
fprintf('Raíz aproximada: %.6f\n', raiz_new); 
fprintf('Iteraciones: %d\n', iter_new); 
fprintf('Cota de error: %.6e\n', error_new); 