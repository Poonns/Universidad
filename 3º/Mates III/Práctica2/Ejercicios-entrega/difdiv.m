function [A,v,pc]=difdiv(x,f)
m=length(x);
    for i=1:m
        A(i,1)=f(i);
    end
    for j=2:m
        for i=j:m
          A(i,j)=(A(i,j-1)-A(i-1,j-1))/(x(i)-x(i-j+1));
        end
    end
v=diag(A);
% calculo del polinomio
    % en forma canonica (pc)
    pc = A(m,m);
    for i=(m-1):-1:1
        pc = conv(pc,poly(x(i)));
        n = length(pc);
        pc(n) = pc(n) + A(i,i);
    end
end