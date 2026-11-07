%parametros iniciales
x0 = 0; 
y0 = 1; 
xf = 1;

%definimos las EDOs (y' = f(x,y))
f1 = @(x,y) (x^2) * y;
f2 = @(x,y) (1 + cos(x)) * y;

%soluciones exactas
y1_exacta = @(x) exp((x.^3)/3);
y2_exacta = @(x) exp(x + sin(x));

%vectores de pasos h a evaluar
h_vals = [0.1, 0.01];

for h = h_vals
    disp(['Resultados para h = ', num2str(h),]);

    %resolucion EDO 1
    [x1, ye1, ~] = ee(f1, x0, y0, h, xf);
    [~, yrk1, ~] = rk4(f1, x0, y0, h, xf);

    %resolucion EDO 2
    [x2, ye2, ~] = ee(f2, x0, y0, h, xf);
    [~, yrk2, ~] = rk4(f2, x0, y0, h, xf);

    %calculo de errores en x=1 (último elemento del vector)
    err_e1 = abs(y1_exacta(1) - ye1(end));
    err_rk1 = abs(y1_exacta(1) - yrk1(end));

    err_e2 = abs(y2_exacta(1) - ye2(end));
    err_rk2 = abs(y2_exacta(1) - yrk2(end));

    disp(['EDO 1 (x=1) -> Error Euler: ', num2str(err_e1), ' | Error RK4: ', num2str(err_rk1)]);
    disp(['EDO 2 (x=1) -> Error Euler: ', num2str(err_e2), ' | Error RK4: ', num2str(err_rk2)]);
    disp(' '); 


    %representacion grafica
    figure('Name', ['Comparativa para h = ', num2str(h)]);

    %grafica EDO 1
    subplot(1, 2, 1);
    plot(x1, ye1, 'r-o', x1, yrk1, 'b-s', x1, y1_exacta(x1), 'k-', 'LineWidth', 1.5);
    title(['EDO 1: y'' = x^2 y (h = ', num2str(h), ')']);
    xlabel('x'); ylabel('y(x)');
    legend('Euler', 'RK4', 'Exacta', 'Location', 'best');
    grid on;

    %grafica EDO 2
    subplot(1, 2, 2);
    plot(x2, ye2, 'r-o', x2, yrk2, 'b-s', x2, y2_exacta(x2), 'k-', 'LineWidth', 1.5);
    title(['EDO 2: y'' = (1+cos(x))y (h = ', num2str(h), ')']);
    xlabel('x'); ylabel('y(x)');
    legend('Euler', 'RK4', 'Exacta', 'Location', 'best');
    grid on;
end
