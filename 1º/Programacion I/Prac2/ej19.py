a = int(input('Introduce el lado a: '))
b = int(input('Introduce el lado b: '))
c = int(input('Introduce el lado c: '))

while not (a + b > c and a + c > b and b + c > a):
    print('No es triangulo')
    a = int(input('Introduce el lado a: '))
    b = int(input('Introduce el lado b: '))
    c = int(input('Introduce el lado c: '))

print('Es triangulo')
if a == b and b == c:
    print('El triangulo es equilatero')
elif (a == b and a != c) or (a == c and  a != b) or (b == c and b != a):
    print('El triangulo es isosceles')
else:
    print('El triangulo es escaleno')