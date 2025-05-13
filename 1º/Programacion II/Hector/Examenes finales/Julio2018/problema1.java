package Hector.Examenes finales;

public class Julio2018 {
    public class Cuenta{
        private String cuenta;
        private String titular;
        private double saldo;
        
        Cuenta(String numCuenta, String titular){
            this.cuenta = numCuenta;
            this.titular = titular;
        }

        String getNumeroCuenta(){
            return this.cuenta;
        }

        String getTitular(){
            return this.titular;
        }

        double getSaldo(){
            return saldo;
        }

        void ingresar(double importe){
            saldo += importe;
        }

        void reintegrar(double importe){
            saldo -= importe;
        }
    }

    public class Sucursal {
        private String codigo; 
        private Cuenta[] cuentas;
        
        public Sucursal(String codigo) {
            this.codigo = codigo;
            this.cuentas = new Cuenta[0]; 
        }

        void añadirCuenta(String numeroCuenta, String titular){
            Cuenta nuevo = new Cuenta(numeroCuenta, titular);
            Cuenta[] vector = new Cuenta[cuentas.length + 1];

            int posicion = 0;
            for(int i = 0; i < cuentas.length; i++){
                if(nuevo.getNumeroCuenta().compareTo(cuentas[i].getNumeroCuenta()) < 0){
                    posicion = i;
                    break;
                }
                vector[i] = cuentas[i];
            }

            vector[posicion] = nuevo;

            for(int j = posicion; j < cuentas.length; j++){
                vector[j+1] = cuentas[j];
            }

            cuentas = vector;
        }

        void asentarOperacion(char tipo, String c1, String c2, double importe){
            int i = -1;
            int inicio = 0;
            int fin = cuentas.length -1;

            while(inicio <= fin){
                int media = (inicio + fin) / 2;
                String actual = cuentas[media].getNumeroCuenta();
                int comparacion = c1.compareTo(actual);

                if(comparacion == 0){
                    i = media;
                    break;
                } else if(comparacion < 0){
                    fin = media -1;
                } else{
                    inicio = media +1;
                }
            }

            if(i == -1){
                return;
            }

            int j = -1;
            if(tipo == 'T'){
                inicio = 0;
                fin = cuentas.length -1;

                while(inicio <= fin){
                    int media = (inicio + fin) / 2;
                    String actual = cuentas[media].getNumeroCuenta();
                    int comparacion = c2.compareTo(actual);

                    if(comparacion == 0){
                        j = media;
                        break;
                    } else if(comparacion < 0){
                        fin = media -1;
                    } else{
                        inicio = media +1;
                    }
                }
                if(j == -1){
                    return;
                }

                switch(tipo){
                    case 'I':
                        cuentas[i].ingresar(importe);
                        break;
                    case 'R':
                        cuentas[i].reintegrar(importe);
                        break;
                    case 'T':
                        cuentas[j].ingresar(importe);
                        cuentas[i].ingresar(importe);
                        break;
                }
            }
        }

        int borrar(String titular){
            int contador = 0;
            for(int i = 0; i < cuentas.length; i++){
                if(titular.equals(cuentas[i].getTitular())){
                    contador += 1;
                }
            }

            Cuenta[] nuevovector = new Cuenta[cuentas.length - contador];

            int j = 0;
            for(int i = 0; i < nuevovector.length; i++){
                if(!titular.equals(cuentas[i].getTitular())){
                    nuevovector[j] = cuentas[i];
                    j++;
                }
            }
            nuevovector = cuentas;
            return contador;
        }
    }
}
