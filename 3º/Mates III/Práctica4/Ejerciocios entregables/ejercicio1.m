syms x;
f1_sym = sin(30*x);
f2_sym = sin(x/30);
a1 = 0; b1 = pi/2;

%a)valores exactos
I1_exacto = double(int(f1_sym, x, a1, b1));
I2_exacto = double(int(f2_sym, x, a1, b1));
fprintf('I1 exacto: %f\n', I1_exacto);
fprintf('I2 exacto: %f\n\n', I2_exacto);

%aproximaciones, cotas y errores para N = 50, 100, 150, 200
N_vals = [50, 100, 150, 200];
f1 = @(x) sin(30*x);
f2 = @(x) sin(x/30);

fprintf('Resultados para I1 = sin(30x):\n');
fprintf('N\t| Método\t| Aproximación\t| Cota de Error\t| Error Real\n');
fprintf('--------------------------------------------------------------------\n');
for N = N_vals
    [vi_R, er_R] = rectangulos(f1, a1, b1, N);
    [vi_T, er_T] = trapecios(f1, a1, b1, N);
    err_R = abs(I1_exacto - vi_R);
    err_T = abs(I1_exacto - vi_T);
    fprintf('%d\t| Rectáng.\t| %f\t| %e\t| %e\n', N, vi_R, er_R, err_R);
    fprintf('%d\t| Trapecio\t| %f\t| %e\t| %e\n', N, vi_T, er_T, err_T);
end

fprintf('\nResultados para I2 = sin(x/30):\n');
fprintf('N\t| Método\t| Aproximación\t| Cota de Error\t| Error Real\n');
fprintf('--------------------------------------------------------------------\n');
for N = N_vals
    [vi_R, er_R] = rectangulos(f2, a1, b1, N);
    [vi_T, er_T] = trapecios(f2, a1, b1, N);
    err_R = abs(I2_exacto - vi_R);
    err_T = abs(I2_exacto - vi_T);
    fprintf('%d\t| Rectáng.\t| %f\t| %e\t| %e\n', N, vi_R, er_R, err_R);
    fprintf('%d\t| Trapecio\t| %f\t| %e\t| %e\n', N, vi_T, er_T, err_T);
end