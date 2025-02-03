a = int(input('Introduce el primer número: '))
b = int(input('Introduce el segundo número: '))
c = int(input('Introduce el tercer número: '))
d = int(input('Introduce el cuarto número: '))

if a < b and a < c and a < d:
    print(f'{a} es el menor')
elif b < a and b < c and b < d:
    print(f'{b} es el menor ')
elif c < a and c < b and c < d:
    print(f'{c} es el menor ')
else:
    print(f'{d} es el menor')
