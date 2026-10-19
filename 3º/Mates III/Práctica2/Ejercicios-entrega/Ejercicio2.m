%apartado 1%
T = [0, 50, 100, 150, 200, 250, 300, 350, 400];
mV = [0.001, 1.4, 5.2, 8.1, 15.8, 20.0, 31.5, 32.0, 45.1];

tabla = difdiv(T, mV);
disp('La tabla es :'); 
disp(tabla);

%apartado 2%
T_fino = linspace(0, 400, 200);
mV_fino = zeros(size(T_fino));

for i = 1:length(T_fino)
    mV_fino(i) = interpdifdiv(T_fino(i), T, mV);
end

figure;
plot(T, mV, 'ro', T_fino, mV_fino, 'b-', 'LineWidth', 1.5);
grid on;
xlabel('Temperatura T (°C)');
ylabel('Señal del Sensor (mV)');
title('Calibración del Sensor (9 nodos)');
legend('Datos originales', 'Polinomio interpolador', 'Location', 'NorthWest');

%apartado 3%
est_25 = interpdifdiv(25, T, mV);
disp('Para T = 25 ºC: ');
disp(est_25);

est_175 = interpdifdiv(175, T, mV);
disp('Para T = 175 ºC: ');
disp(est_175);

est_375 = interpdifdiv(375, T, mV);
disp('Para T = 375 ºC: ');
disp(est_375);

%apartado 4%
T5 = [0, 100, 200, 300, 400];
mV5 = [0.001, 5.2, 15.8, 31.5, 45.1]; 

tabla5 = difdiv(T5, mV5);


est5_25 = interpdifdiv(25, T5, mV5);
disp('Para T = 25 ºC: ');
disp(est5_25);
est5_175 = interpdifdiv(175, T5, mV5);
disp('Para T = 175 ºC: ');
disp(est5_175);
est5_375 = interpdifdiv(375, T5, mV5);
disp('Para T = 375 ºC: ');
disp(est5_375);
