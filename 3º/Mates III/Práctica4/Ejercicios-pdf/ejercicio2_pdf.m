%definicion de funciones y parametros
f = @(x) log(1+x);
a = 0;
b = 2;

%valor de la integral
I_exacto = integral(f, a, b);
fprintf('1. Valor exacto de la integral: %f\n\n', I_exacto);

%regla del Rectángulo (N=50 y N=100)
[vi_R50, er_R50] = rectangulos(f, a, b, 50);
error_real_R50 = abs(I_exacto - vi_R50);

[vi_R100, er_R100] = rectangulos(f, a, b, 100);
error_real_R100 = abs(I_exacto - vi_R100);

fprintf('2. Método del Rectángulo\n');
fprintf('   N=50  -> Aprox: %f, Cota: %e, Error Real: %e\n', vi_R50, er_R50, error_real_R50);
fprintf('   N=100 -> Aprox: %f, Cota: %e, Error Real: %e\n\n', vi_R100, er_R100, error_real_R100);

%cociente para Rectángulo
cociente_R = error_real_R50 / error_real_R100;
fprintf('3. Cociente Rectángulo (E50/E100): %f\n', cociente_R);
fprintf('   Al duplicar N, el paso h se reduce a la mitad. Como el método del rectángulo\n');
fprintf('   es de orden 1 (el error es proporcional a h), el error se divide por 2.\n');
fprintf('   El cociente cercano a 2 confirma matemáticamente este orden.\n\n');

%regla del Trapecio (N=50 y N=100)
[vi_T50, er_T50] = trapecios(f, a, b, 50);
error_real_T50 = abs(I_exacto - vi_T50);

[vi_T100, er_T100] = trapecios(f, a, b, 100);
error_real_T100 = abs(I_exacto - vi_T100);

fprintf('4. Método del Trapecio\n');
fprintf('   N=50  -> Aprox: %f, Cota: %e, Error Real: %e\n', vi_T50, er_T50, error_real_T50);
fprintf('   N=100 -> Aprox: %f, Cota: %e, Error Real: %e\n\n', vi_T100, er_T100, error_real_T100);

%cociente para Trapecio
cociente_T = error_real_T50 / error_real_T100;
fprintf('5. Cociente Trapecio (E50/E100): %f\n', cociente_T);
fprintf('   El trapecio es un método de orden 2 \n');
fprintf('   Al dividir h por 2, el error se divide por 2^2 \n');
fprintf('   El cociente cercano a 4 confirma este comportamiento teórico\n');