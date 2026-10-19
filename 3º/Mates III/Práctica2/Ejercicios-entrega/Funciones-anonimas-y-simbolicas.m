%% nota sobre funciones anónimas y simbólicas y la derivacion e integración simbólica en matlab
% funciones anónimas: son funciones como las que podemos definir en un fichero .m
% por ejemplo
clc
syms x;
clear fa
fa=@(x) sin(x/2) + x
% pueden usarse directamente pasándoles un valor como cualquier otra función interna de matlab
fa(pi)
% funciones simbólicas: son mas bien una representacion textual de la funcion
clear fs
fs=sin(x/2) + x
% No pueden usarse directamente
%fs(pi)
% Para obtener el resultado, se debe usar primero subs:
subs(fs,x,pi)
% y con caracter general, despues vpa
vpa(subs(fs,x,pi))

%% convertir una función anónima en simbólica
% Una función anónima puede trabajar tanto con argumentos numéricos como
% con variables simbólicas
% su usamos una variable numérica, obtenemos un resultado numérico
clc
t=1;
fa(t)

% Pero cuando evaluamos una función anónima con un argumento simbólico obtenemos una función simbólica
syms t;
fa(t)

%% Convertir una función simbólica en anónima
% para la conversion inversa, disponemos de la función "matlabFunction"
fs
FS=matlabFunction(fs)
%fs(pi)
FS(pi)


%% diff proporciona la derivada de una función simbólica (y devuelve también una f. simbolica)
% pero falla con las funciones anonimas:
%dfa = diff(fa)
dfs = diff(fs)
% sin embargo, como las funciones anónimas pueden usar variables simbólicas, podemos construir una funcion simbólica a partir de una anónima y usar luego diff
syms t
fa(t)
diff(fa(t))

%% int devuelve la primitiva de una funcion simbólica
int(fs)

%% mínimo (y máximo ) numéricos de una función en un intervalo
% la funcion de matlab:  fminbnd(fun,a,b)
% devuelve el valor de x que minimiza la función en el intervalo [a,b]
% la función que se pasa debe ser una función anónima
% se trata de una minimización numtrica y en consecuecia, sujeta a errores de redondeo
syms x;
h=@(x)cos(x)
fminbnd(h,0,2*pi)
g=@(x)x^2
fminbnd(g,-1,1)

% para poder usar una función simbólica, debemos convertirla primero a anónima
g=cos(x)
%fminbnd(g,0,2*pi)
G=matlabFunction(g)
fminbnd(G,0,2*pi)

% No existe la función equivalente para encontrar el máximo, pero si cambiamos de signo la función
% los máximos de la función inicial se convierten en mínimos en la función cambiada de signo:

% Ejemplo: encontrar el máximo de f(x) = x^2 en el intervalo [-1, 1]
% definimos g(x) = -f(x):
g=@(x) -x^2 ;
disp("Valor de x que minimiza g(x) = - x^2 y maximiza f(x) = x^2")
fminbnd(g,-1,1)

