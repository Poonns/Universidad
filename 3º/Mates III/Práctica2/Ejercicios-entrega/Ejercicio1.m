%apartado 1%
x = linspace(0, 3, 7);
fx = 7 * sin(x.^2) / 21;

%apartado 2%
tabla = difdiv(x, fx);
disp('La tabla es:');
disp(tabla);

% apartado 3%
syms z;
f = 7*sin(z^2) / 21;
derivada = diff(f, 7);

puntos = linspace(0, 3, 1000);
max_derivada = max(double(abs(subs(derivada, z, puntos))));
disp('Valor máximo de la derivada 7:');
disp(max_derivada);

a = 2.3;
cotaError = (max_derivada / factorial(7)) * prod(abs(a - x));
disp('Cota teórica de error:');
disp(cotaError);

%apartado 4%
resultado_interpolacion = interpdifdiv(a, x, fx);
disp('Valor interpolado en x = 2.3:');
disp(resultado_interpolacion);

valor = double(subs(f, z, a));
error = abs(valor - resultado_interpolacion);

disp('Valor exacto f(2.3):');
disp(valor);
disp('Error absoluto obtenido:');
disp(error);
