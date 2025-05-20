package Hector.Examenes

import java.util.Scanner;
import java.io.File;
import java.io.FileNotFoundException;


public class problema2 {
    class Cuentas{
        private class Nodo{
            String nombre;
            double cantidad;
            Nodo sig;

            Nodo(String nombre, double cantidad, Nodo sig){
                this.nombre = nombre;
                this.cantidad = cantidad;
                this.sig = sig;
            }
        }
        private Nodo primero;

        private void insertar(String nombre, double cantidad){
            for(Nodo actual = primero; actual != null; actual = actual.sig){
                if(actual.nombre.equals(nombre)){
                    actual.cantidad += cantidad;
                    return;
                }
            }

            Nodo nuevo = new Nodo(nombre, cantidad, primero);
            primero = nuevo;
        }
        
        private void convertirASaldos(){
            double cantidad = 0;
            int personas = 0;

            for(Nodo actual = primero; actual != null; actual = actual.sig){
                cantidad += actual.cantidad;
                personas ++;
            }

            double media = cantidad / personas;

            for(Nodo actual = primero; actual != null; actual = actual.sig){
                actual.cantidad -= media;
            }
        }

        Cuentas(String nombreFichero) throws FileNotFoundException{
            Scanner entrada = new Scanner(new File(nombreFichero));

            while(entrada.hasNextLine()){

                String[] partes = entrada.nextLine().split(" ");

                if(partes.length == 2){
                    String saldo = partes[0];
                    String nombre = partes[1];
                    insertar(nombre, 0);
                }
            }

            entrada.close();
            convertirASaldos();
        }

        public void asentarPago(String pagador, String receptor, double cantidad){
            Nodo nodopagador = null, nodoreceptor = null;
            Nodo antpagador = null, antreceptor = null;
            Nodo anterior = null;
            
            for(Nodo actual = primero; actual != null; actual = actual.sig){
                if(actual.nombre.equals(pagador)){
                    actual.cantidad += cantidad;
                    nodopagador = actual;
                    antpagador = anterior;
                }else if(actual.nombre.equals(receptor)){
                    actual.cantidad -= cantidad;
                    nodoreceptor = actual;
                    antreceptor = anterior;
                }
            }
            if(nodopagador != null && nodopagador.cantidad == 0){
                if(antpagador == null){
                    primero = nodopagador.sig;
                } else {
                    antpagador.sig = nodopagador.sig;
                }
            }

            if(nodoreceptor != null && nodoreceptor.cantidad == 0){
                if(antreceptor != null){
                    primero = nodoreceptor.sig;
                } else {
                    antreceptor.sig = nodoreceptor.sig;
                }
            }
        }
    }
}
