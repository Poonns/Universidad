a = int(input('Introduce el primer número: '))
b = int(input('Introduce el segundo número: '))
c = int(input('Introduce el tercer número: '))

if a < b and a < c:
    print(f'{a} es el menor')
elif b < a and b < c:
    print(f'{b} es el menor ')
else:
    print(f'{c} es el menor')
