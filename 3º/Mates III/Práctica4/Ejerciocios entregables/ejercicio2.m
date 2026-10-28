syms x

f3_sym = exp(-x^2);
a2 = 0;
b2 = 1;

%a y b)primitiva y valor exacto
primitiva = int(f3_sym, x);
disp('Primitiva calculada por MATLAB:');
disp(primitiva);

I3_exacto = double(int(f3_sym, x, a2, b2));
fprintf('\nValor exacto de la integral: %f\n\n', I3_exacto);

%c)aproximaciones N=40
f3 = @(x) exp(-x.^2);
N3 = 40;

[vi_R3, er_R3] = rectangulos(f3, a2, b2, N3);
[vi_T3, er_T3] = trapecios(f3, a2, b2, N3);
err_R3 = abs(I3_exacto - vi_R3);
err_T3 = abs(I3_exacto - vi_T3);

fprintf('Resultados para exp(-x^2) con N = 40:\n');
fprintf('Método\t\t| Aproximación\t| Cota de Error\t| Error Real\n');
fprintf('--------------------------------------------------------------------\n');
fprintf('Rectángulos\t| %f\t| %e\t| %e\n', vi_R3, er_R3, err_R3);
fprintf('Trapecios\t| %f\t| %e\t| %e\n', vi_T3, er_T3, err_T3);