public class Controlador implements ActionListener{
    private Vista vista;
    private Modelo modelo;

    public Controlador(Vista v, Modelo m){
        vista = v;
        modelo = m;

        vista.getBoton1().addActionListener(this);
        vista.getBoton2().addActionListener(this);
        vista.getBoton3().addActionListener(this);

        vista.getItems().addActionListener(this);

        for(int i = 0; i <= 15; i++){
            vista.getBotonesPanel()[i].addActionListener(this);
        }
    }
    @Override
    public void actionPerformed(ActionEvent e){
        String texto = "";
        
        if(e.getSource() == vista.getItems()){
            texto = (String) vista.getItems().getSelectedItem();
        } else {
            texto = e.getActionCommand();
        }

        modelo.setUltimaAccion(texto);
        vista.setEtiqueta(texto);
    }
}