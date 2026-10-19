f2 = @(x) (x.^3)/2 - 6*x + 5;
fd2 = @(x) (3*x.^2)/2 - 6;


val_a = f2(-4);
val_b = f2(-3);

fprintf('Evaluación en los extremos del intervalo [0, 1]:\n');
fprintf('f1(0) = %.4f \n', val_a);
fprintf('f1(1) = %.4f \n', val_b);
disp('Existe un cambio de signo, por lo que hay al menos una raíz.'); 

x = linspace(-4, -3, 100); 
y = f2(x);
x0 = -3.5;

plot(x, y) 
yline(0, '--k') 
grid on 
xlabel('Eje X') 
ylabel('Eje Y') 
title('Raíz de f_2(x) = (x^3)/2 - 6*x + 5; en [-4,-3]') 
legend('f_2(x)', 'y = 0') 

tol = 1e-5;
maxiteraciones = 100;
%Calculo Via Biseccion%
[raiz_bis, iter_bis, error_bis] = biseccion(-4, -3, f2, maxiteraciones, tol);
fprintf('Raíz aproximada: %.6f\n', raiz_bis);
fprintf('Iteraciones: %d\n', iter_bis); 
fprintf('Cota de error: %.6e\n', error_bis); 

%Calculo Via Newton%
[raiz_new, iter_new, error_new] = newton(x0, f2, fd2, maxiteraciones, tol); 
fprintf('Raíz aproximada: %.6f\n', raiz_new); 
fprintf('Iteraciones: %d\n', iter_new); 
fprintf('Cota de error: %.6e\n', error_new); 