f1 = @(x) cos(x) - x;
fd1 = @(x) -sin(x) - 1;


val_a = f1(0);
val_b = f1(1);

fprintf('Evaluación en los extremos del intervalo [0, 1]:\n');
fprintf('f1(0) = %.4f \n', val_a);
fprintf('f1(1) = %.4f \n', val_b);
disp('Existe un cambio de signo, por lo que hay al menos una raíz.'); 

x = linspace(0, 1, 100); 
y = f1(x);
x0 = 0.5;

plot(x, y) 
yline(0, '--k') 
grid on 
xlabel('Eje X') 
ylabel('Eje Y') 
title('Raíz de f_1(x) = cos(x) - x en [0,1]') 
legend('f_1(x)', 'y = 0') 

tol = 1e-5;
maxiteraciones = 100;
%Calculo Via Biseccion%
[raiz_bis, iter_bis, error_bis] = biseccion(0, 1, f1, maxiteraciones, tol);
fprintf('Raíz aproximada: %.6f\n', raiz_bis);
fprintf('Iteraciones: %d\n', iter_bis); 
fprintf('Cota de error: %.6e\n', error_bis); 

%Calculo Via Newton%
[raiz_new, iter_new, error_new] = newton(x0, f1, fd1, maxiteraciones, tol); 
fprintf('Raíz aproximada: %.6f\n', raiz_new); 
fprintf('Iteraciones: %d\n', iter_new); 
fprintf('Cota de error: %.6e\n', error_new); 