a = float(input('Introduce el valor del cateto a: '))
b = float(input('Introduce el valor del cateto b: '))
c = float(input('Introduce el valor del cateto c: '))

if (a == b and a != c) or (a == c and  a != b) or (b == c and b != a):
    print('Es un triangulo isosceles')
else:
    print('No es un triangulo isosceles')