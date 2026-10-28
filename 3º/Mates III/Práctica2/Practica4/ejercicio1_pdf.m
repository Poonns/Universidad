%definicion de funcion y parametros
f = @(x) cos(x);
a = 0;
b = pi/2;
N = 50;

%valor exacto de la integral
I_exacto = integral(f, a, b); 
fprintf('1. Valor exacto de la integral: %f\n\n', I_exacto);

%regla del Rectángulo (N=50)
[vi_R, er_R] = rectangulos(f, a, b, N);
error_real_R = abs(I_exacto - vi_R);
fprintf('2. Método del Rectángulo (N=50)\n');
fprintf('   Aproximación: %f\n', vi_R);
fprintf('   Cota teórica: %e\n', er_R);
fprintf('   Error real:   %e\n\n', error_real_R);

%regla del Trapecio (N=50)
[vi_T, er_T] = trapecios(f, a, b, N);
error_real_T = abs(I_exacto - vi_T);
fprintf('3. Método del Trapecio (N=50)\n');
fprintf('   Aproximación: %f\n', vi_T);
fprintf('   Cota teórica: %e\n', er_T);
fprintf('   Error real:   %e\n\n', error_real_T);

%comparación teórica
fprintf('4. Comparación y Justificación\n');
fprintf('   El método del trapecio es más preciso. Según la teoría ,\n');
fprintf('   el error del rectángulo es proporcional a "h" , mientras\n');
fprintf('   que el del trapecio es proporcional a "h^2"\n');
fprintf('   Como el paso h es menor que 1, elevarlo al cuadrado lo hace mucho más pequeño\n');