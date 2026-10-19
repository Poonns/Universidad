x = [-1.0, -0.5, 0.0, 0.5, 1.0];
fx = [0.500000, 0.800000, 1.000000, 0.800000, 0.500000];
a = 0.3;

tabla = difdiv(x, fx);
resultado_interpolacion = interpdifdiv(a, x, tabla);
disp('Aprximacion de grado 4:');
disp(resultado_interpolacion);

syms z;
f1 = 1 / (1 + z^2);
derivada = diff(f1, 5);

puntos = linspace(-1.0, 1.0, 1000);
max_derivada = (max(double(abs(subs(derivada, z, puntos)))))

cotaError = (max_derivada / factorial(5)) * prod(abs(a - x));
disp('La cota de error teoríca es:');
disp(cotaError);

error = abs(1 / (1 + 0.3^2) - resultado_interpolacion);
disp('El error es:');
disp(error);