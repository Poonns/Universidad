
%a) y b) Solución numérica RK4 y cálculo de errores
f2 = @(x,y) x * y;
x0 = 0;
y0 = 1;
xf = 1;
h_vals = [0.1, 0.01, 0.001];

y2_exacta = @(x) exp((x^2)/2);

disp(' Tabla de errores y cocientes con RK4');
disp('(h)   |   Error exacto (x=1)   |   Cociente (Error / h^4)');
disp('--------------------------------------------------------');
for h = h_vals
    [x2, y2_num, ~] = rk4(f2, x0, y0, h, xf);
    error2 = abs(y2_exacta(1) - y2_num(end));
    cociente = error2 / (h^4);
    disp([num2str(h), '      | ', num2str(error2), '       | ', num2str(cociente)]);
end
