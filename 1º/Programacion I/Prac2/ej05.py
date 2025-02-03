a = float(input('Introduce el valor del cateto a: '))
b = float(input('Introduce el valor del cateto b: '))
c = float(input('Introduce el valor del cateto c: '))

if a != b and b != c and a != c:
    print('El triangulo es escaleno')
else:
    print('El triangulo no es escaleno')