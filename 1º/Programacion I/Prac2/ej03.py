a = float(input('Introduce el valor del cateto a: '))
b = float(input('Introduce el valor del cateto b: '))
c = float(input('Introduce el valor del cateto c: '))

if a == b and b == c:
    print('El triangulo es equilatero')
else:
    print('El triangulo no es equilatero')