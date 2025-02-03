a = float(input('Introduce el valor del cateto a: '))
b = float(input('Introduce el valor del cateto b: '))
c = float(input('Introduce el valor del cateto c: '))

if a + b > c and a + c > b and b + c > a:
    print('Es un triangulo')
else:
    print('No es un triangulo')