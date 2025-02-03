a = int(input('Introduce un numero: '))
b = int(input('Introduce otro numero: '))

if a * b < 0:
    print('El valor es negativo')
elif a * b == 0:
    print('El resultado es nulo')
else:
    print('El resultado es positivo')
