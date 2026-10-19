x = [0.90, 1.00, 1.10, 1.20];
fx = [2.459603, 2.718282, 3.004166, 3.320117];

%le pasamos los vectores a la funcion difdiv para que nos cree la tabla%
tabla = difdiv(x, fx);
disp('Tabla de diferencias divididas:');
disp(tabla);

%evaluamos el polinomio interpolador de grado 3 en x = 1.05%
a = 1.05;
resultado_interpolacion = interpdifdiv(a, x, tabla);
disp('Resultado al evaluar la funcion:')
disp(resultado);

%calculamos la cota de error, para ello debemos calcular la derivada en n+1%
syms z;
f1 = exp(z);
derivada = diff(f1, 4);

I = [0.90, 1.20];
cotaError = double((subs(derivada, z, 1.20) / factorial(4)) * prod(abs(a-x))); %usamos subs para sustituir el valor en x para la derivada%
disp('La cota de error teoríca es:');
disp(cotaError);

%calculamos el error real%
error = abs(exp(1.05) - resultado_interpolacion);
disp('El error real es:');
disp(error);