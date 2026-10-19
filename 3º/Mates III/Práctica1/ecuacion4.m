f4 = @(x) log(x+2) - x/2;
fd4 = @(x) 1./(x+2) - 1/2;

val_a = f3(3);
val_b = f3(4);

fprintf('Evaluación en los extremos del intervalo [3, 4]:\n');
fprintf('f4(3) = %.4f \n', val_a);
fprintf('f4(4) = %.4f \n', val_b);
disp('Existe un cambio de signo, por lo que hay al menos una raíz.'); 

x = linspace(3, 4, 100); 
y = f4(x);
x0 = 3.5;

plot(x, y) 
yline(0, '--k') 
grid on 
xlabel('Eje X') 
ylabel('Eje Y') 
title('Raíz de f_4(x) = log(x+2) - x/2; en [3,4]') 
legend('f_4(x)', 'y = 0') 

tol = 1e-5;
maxiteraciones = 100;
%Calculo Via Biseccion%
[raiz_bis, iter_bis, error_bis] = biseccion(3, 4, f4, maxiteraciones, tol);
fprintf('Raíz aproximada: %.6f\n', raiz_bis);
fprintf('Iteraciones: %d\n', iter_bis); 
fprintf('Cota de error: %.6e\n', error_bis); 

%Calculo Via Newton%
[raiz_new, iter_new, error_new] = newton(x0, f4, fd4, maxiteraciones, tol); 
fprintf('Raíz aproximada: %.6f\n', raiz_new); 
fprintf('Iteraciones: %d\n', iter_new); 
fprintf('Cota de error: %.6e\n', error_new); 