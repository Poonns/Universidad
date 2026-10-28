function [vi, er] = rectangulos(f, a, b, n)
    h = (b - a) / n;
    
    %genercion de nodos desde a hasta (b-h)
    x_k = a : h : (b - h); 
    
    %valor aprox integral
    vi = h * sum(f(x_k));
    
    %calculo de cota teorica de error usando variables simbolicas
    syms x;
    f_sym = f(x);
    df = diff(f_sym, x, 1);
    
    %evaluamos la derivada para encontrar el maximo
    puntos = linspace(a, b, 1000);
    max_df = max(abs(double(subs(df, x, puntos))));
    
    %cota de error
    er = ((b - a) / 2) * h * max_df;
end