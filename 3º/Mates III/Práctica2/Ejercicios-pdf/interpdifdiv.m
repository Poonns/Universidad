function yf=interpdifdiv(xf,x,y)
    m=length(x);
    [A,v]=difdiv(x,y);
    yf=v(1);
    for i=2:m
        yf=yf+v(i)*prod(xf-x(1:i-1));
    end
end    