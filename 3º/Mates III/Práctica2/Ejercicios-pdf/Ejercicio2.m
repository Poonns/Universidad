%grado 1%
x1 = [1.20, 1.40];
fx1 = [0.182322, 0.336472];
a = 1.30;
tabla1 = difdiv(x1, fx1);
resultado_interpolacion1 = interpdifdiv(a, x1, tabla1);
disp('Aprximacion de grado 1:');
disp(resultado_interpolacion1);

syms z;
f1 = log(z);
derivada1 = diff(f1, z , 2);

%calculamos 1000 puntos para saber cual es el maximo en el intervalo%
puntos =linspace(1.20, 1.40, 1000);
maximo_derivada1 = max(double(abs(subs(derivada1, z, puntos))));

cotaError1 = (maximo_derivada1 / factorial(2) * prod(abs(a - x1)));
disp('La cota teoríca de error de f1 es:')
disp(cotaError1);

error1 = abs(log(1.30) - resultado_interpolacion1);
disp('El error es:');
disp(error1);

%grado 2%
x2 = [1.00, 1.20, 1.40];
fx2 = [0.000000, 0.182322, 0.336472];

tabla2 = difdiv(x2, fx2);
resultado_interpolacion2 = interpdifdiv(a, x2, tabla2);
disp('Aprximacion de grado 2:');
disp(resultado_interpolacion2);

derivada2 = diff(f1, z , 3);

%calculamos 1000 puntos para saber cual es el maximo en el intervalo%
puntos =linspace(1.00, 1.40, 1000);
maximo_derivada2 = max(double(abs(subs(derivada2, z, puntos))));

cotaError2 = (maximo_derivada2 / factorial(3) * prod(abs(a - x2)));
disp('La cota teoríca de error de f2 es:')
disp(cotaError2);

error2 = abs(log(1.30) - resultado_interpolacion2);
disp('El error es:');
disp(error2);

%grado 3%
x3 = [1.00, 1.20, 1.40, 1.60];
fx3 = [0.000000, 0.182322, 0.336472, 0.470004];

tabla3 = difdiv(x3, fx3);
resultado_interpolacion3 = interpdifdiv(a, x3, tabla3);
disp('Aprximacion de grado 3:');
disp(resultado_interpolacion3);

derivada3 = diff(f1, z , 4);

%calculamos 1000 puntos para saber cual es el maximo en el intervalo%
puntos =linspace(1.00, 1.60, 1000);
maximo_derivada3 = max(double(abs(subs(derivada3, z, puntos))));

cotaError3 = (maximo_derivada3 / factorial(4) * prod(abs(a - x3)));
disp('La cota teoríca de error de f3 es:')
disp(cotaError3);

error3 = abs(log(1.30) - resultado_interpolacion3);
disp('El error es:');
disp(error3);