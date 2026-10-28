function [vi, er] = trapecios(f, a, b, n)
    h = (b - a) / n;
    
    %nodos intermedios desde (a+h) hasta (b-h)
    x_k = (a + h) : h : (b - h); 
    
    %formula del trapecio compuesto
    vi = (h / 2) * (f(a) + 2 * sum(f(x_k)) + f(b));
    
    %calculo de la cota teorica de error en la segunda derivada
    syms x;
    f_sym = f(x);
    d2f = diff(f_sym, x, 2); 
    
    %evaluamos el intervalo para encontrar el maximo
    puntos = linspace(a, b, 1000);
    max_d2f = max(abs(double(subs(d2f, x, puntos))));
    
    %cota de error
    er = ((b - a) / 12) * (h^2) * max_d2f;
end