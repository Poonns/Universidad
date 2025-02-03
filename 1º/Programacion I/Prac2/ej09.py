a = float(input('Introduce el valor del cateto a: '))
b = float(input('Introduce el valor del cateto b: '))
c = float(input('Introduce el valor del cateto c: '))

if (a ** 2) + (b ** 2) == (c ** 2):
    print('El triangulo es rectangulo')
else:
    print('El triangulo no es rectangulo')