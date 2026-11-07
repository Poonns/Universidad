%a) y c) solucion numérica Euler y cálculo de errores
f1 = @(x,y) x - 1 - 3*y;
x0 = 0; 
y0 = 2; 
xf = 1;
h_vals = [0.1, 0.01, 0.001];

%solucion exacta con la K calculada
K = 22/9;
y1_exacta = @(x) K * exp(-3*x) + (1/3)*x - 4/9;

disp('Tabla de errores absolutos con Euler:');
disp('(h)  |  Error exacto en x=1');
disp('-----------------------------------');
for h = h_vals
    [x1, y1_num, ~] = ee(f1, x0, y0, h, xf);
    error1 = abs(y1_exacta(1) - y1_num(end));
    disp([num2str(h), '      | ', num2str(error1)]);
end
disp(' ');
